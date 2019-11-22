package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.ibm.com/ZaaS/ds8k-exporter/utils"

	// "github.com/tidwall/gjson"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	prefix          = "ds8k_"
	defaultEnabled  = true
	defaultDisabled = false
	ds8KAPIPort     = "8452"
)

var (
	scrapeDurationDesc        *prometheus.Desc
	scrapeSuccessDesc         *prometheus.Desc
	requestErrors             *prometheus.Desc
	authTokenCacheCounterHit  *prometheus.Desc
	authTokenCacheCounterMiss *prometheus.Desc
	authTokenCache            sync.Map
	requestErrorCount         int = 0
	authTokenMiss             int = 0
	authTokenHit              int = 0
	factories                     = make(map[string]func() (Collector, error))
	collectorState                = make(map[string]*bool)
)

// DS8kCollector implements the prometheus.Collecotor interface
type DS8kCollector struct {
	targets    []utils.Targets
	location   string
	Collectors map[string]Collector
}

func init() {
	scrapeDurationDesc = prometheus.NewDesc(prefix+"collector_duration_seconds", "Duration of a collector scrape for one resource", []string{"target"}, nil) // metric name, help information, Arrar of defined label names, defined labels
	scrapeSuccessDesc = prometheus.NewDesc(prefix+"collector_success", "Scrape of resource was sucessful", []string{"target"}, nil)
	requestErrors = prometheus.NewDesc(prefix+"request_errors_total", "Errors in request to the DS8K Exporter", []string{"target"}, nil)
	authTokenCacheCounterHit = prometheus.NewDesc(prefix+"authtoken_cache_counter_hit", "Count of authtoken cache hits", []string{"target"}, nil)
	authTokenCacheCounterMiss = prometheus.NewDesc(prefix+"authtoken_cache_counter_miss", "Count of authtoken cache misses", []string{"target"}, nil)
}

func registerCollector(collector string, isDefaultEnabled bool, factory func() (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}
	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default is %s). Use --no-collector.%s to disable it.", collector, helpDefaultState, collector)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)
	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Bool()
	collectorState[collector] = flag
	factories[collector] = factory
}

// newDS8kCollector creates a new DS8k Collector.
func NewDS8kCollector(targets []utils.Targets, location string) (*DS8kCollector, error) {
	collectors := make(map[string]Collector)
	// log.Infof("Enabled collectors:")
	for key, enabled := range collectorState {
		if *enabled {
			// log.Infof(" - %s", key)
			collector, err := factories[key]()
			if err != nil {
				return nil, err
			}
			collectors[key] = collector
		}
	}
	return &DS8kCollector{targets, location, collectors}, nil
}

// Describe implements the Prometheus.Collector interface.
func (c DS8kCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeSuccessDesc
	ch <- scrapeDurationDesc
	ch <- requestErrors
	ch <- authTokenCacheCounterHit
	ch <- authTokenCacheCounterMiss

	for _, col := range c.Collectors {
		col.Describe(ch)
	}
}

// Collect implements the Prometheus.Collector interface.
func (c DS8kCollector) Collect(ch chan<- prometheus.Metric) {

	hosts := c.targets
	wg := &sync.WaitGroup{}
	wg.Add(len(hosts))
	for _, h := range hosts {
		go c.collectForHost(h, ch, wg)
	}
	wg.Wait()
}

func (c *DS8kCollector) collectForHost(host utils.Targets, ch chan<- prometheus.Metric, wg *sync.WaitGroup) {
	defer wg.Done()
	start := time.Now()
	success := 0
	ds8kClient := utils.DS8kClient{
		UserName:  host.Userid,
		Password:  host.Password,
		IpAddress: host.IpAddress,
		Location:  c.location,
	}

	defer func() {
		ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds(), ds8kClient.IpAddress)
		ch <- prometheus.MustNewConstMetric(requestErrors, prometheus.CounterValue, float64(requestErrorCount), ds8kClient.IpAddress)
		ch <- prometheus.MustNewConstMetric(authTokenCacheCounterMiss, prometheus.CounterValue, float64(authTokenMiss), ds8kClient.IpAddress)
		ch <- prometheus.MustNewConstMetric(authTokenCacheCounterHit, prometheus.CounterValue, float64(authTokenHit), ds8kClient.IpAddress)
		ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, float64(success), ds8kClient.IpAddress)
	}()
	// Need to get rid of the goto cheat, replacing with a for loop, and ensureing it has backoff and a short circuit
	lc := 1
	for lc < 4 {
		log.Debugf("Looking for cached Auth Token for %s", host.IpAddress)
		result, ok := authTokenCache.Load(host.IpAddress)
		if !ok {
			log.Debug("Authtoken not found in cache.")
			log.Debugf("Retrieving authToken for %s", host.IpAddress)
			// get our authtoken for future interactions
			authtoken, err := ds8kClient.RetriveAuthToken()
			if err != nil {
				log.Errorf("Error getting auth token for %s, the error was %v", host.IpAddress, err)
				requestErrorCount++
				success = 0
				return
			}
			authTokenCache.Store(host.IpAddress, authtoken)
			result, _ := authTokenCache.Load(host.IpAddress)
			ds8kClient.AuthToken = result.(string)
			authTokenMiss++
			success = 1
		} else {
			log.Debugf("Authtoken pulled from cache for %s", host.IpAddress)
			ds8kClient.AuthToken = result.(string)
			authTokenHit++
			success = 1
		}
		//test to make sure that our auth token is good
		// if not delete it and loop back
		validateURL := "https://" + host.IpAddress + ":8452/api/v1/systems"
		_, err := ds8kClient.CallDS8kAPI(validateURL)
		if err != nil {
			authTokenCache.Delete(host.IpAddress)
			log.Infof("\nInvalidating authToken for %s, re-requesting authtoken....", host.IpAddress)
			lc++
		} else {
			//We have a valid auth token, we can break out of this loop
			break
		}
	}
	if lc >= 4 {
		// looped and failed multiple times, so need to go further
		log.Errorf("Error getting auth token for %s, please check network or username and password.", host.IpAddress)
		requestErrorCount++
		success = 0
		return
	}
	for _, col := range c.Collectors {
		col.Collect(ds8kClient, ch)
	}

}

// Collector is the interface a collector has to implement.
//Collector collects metrics from ds8k using rest api
type Collector interface {
	//Describe describes the metrics
	Describe(ch chan<- *prometheus.Desc)

	//Collect collects metrics from DS8K RESTful API
	// Collect(client utils.DS8kClient, ch chan<- prometheus.Metric, labelvalues []string) error
	Collect(client utils.DS8kClient, ch chan<- prometheus.Metric)
}
