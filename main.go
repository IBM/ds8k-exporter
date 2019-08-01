package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/ds8k-exporter/collector"
	"github.com/ds8k-exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile             = kingpin.Flag("config.file", "Path to configuration file.").Default("ds8k.yaml").String()
	metricsPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	listenAddress          = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9710").String()
	disableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).").Bool()
	hosts                  = kingpin.Flag("web.targets", "Hosts to scrape").String()
	username               = kingpin.Flag("web.user", "Username to use when connecting to DS8K RESTful API").String()
	passwd                 = kingpin.Flag("web.passwd", "Passwd to use when connecting to DS8K RESTful API").String()
	// maxRequests            = kingpin.Flag("web.max-requests", "Maximum number of parallel scrape requests. Use 0 to disable.").Default("40").Int()
	location        = kingpin.Flag("location", "The location or timezone of the storage device, for example: America/New_York").Default("").String()
	cfg             *utils.Config
	enableCollector bool = true
)

type handler struct {
	// exporterMetricsRegistry is a separate registry for the metrics about the exporter itself.
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	// maxRequests             int
}

func main() {

	// Parse flags.
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("ds8k_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	//Bail early if the config is bad.
	log.Infoln("Loading config from", *configFile)
	var err error
	if *location == "" {
		log.Fatalln("Please input the location of ds8k devices.")
	}

	cfg, err = utils.GetConfig(*configFile)

	if err != nil {
		log.Fatalf("Error parsing config file: %s", err)
	}

	log.Infoln("Starting ds8k_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	//Launch http services
	// http.HandleFunc(*metricsPath, handlerMetricRequest)
	http.Handle(*metricsPath, newHandler(!*disableExporterMetrics))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>ds8k exporter</title></head>
			<body>
				<h1>ds8k exporter</h1>
				<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
		</html>`))
	})
	log.Infof("Listening for %s on %s\n", *metricsPath, *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func targetsForRequest(r *http.Request) ([]utils.Targets, error) {
	reqTarget := r.URL.Query().Get("target")
	// var targets []string
	if reqTarget == "" {
		// targets = strings.Split(*hosts, ",")
		return cfg.Targets, nil
	}

	for _, t := range cfg.Targets {
		if t.IpAddress == reqTarget {
			// return []string{t}, nil
			return []utils.Targets{t}, nil
		}
	}

	return nil, fmt.Errorf("the target '%s' is not defined in the configuration file", reqTarget)
}

func newHandler(includeExporterMetrics bool) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		// maxRequests:             maxRequests,
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
		)
	}

	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	targets, err := targetsForRequest(r)
	if err != nil {
		http.Error(w, err.Error(), 400)
	}
	handler, err := h.innerHandler(targets...)
	if err != nil {
		log.Warnln("Couldn't create  metrics handler:", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Couldn't create  metrics handler: %s", err)))
		return
	}
	handler.ServeHTTP(w, r)
}

func (h *handler) innerHandler(targets ...utils.Targets) (http.Handler, error) {
	registry := prometheus.NewRegistry()
	dsc, err := collector.NewDS8kCollector(targets, *location) //new a DS8k Collector
	if err != nil {
		log.Fatalf("Couldn't create collector: %s", err)
	}
	if enableCollector == true {
		log.Infof("Enabled collectors:")
		for n := range dsc.Collectors {
			log.Infof(" - %s", n)
		}
		enableCollector = false
	}

	if err := registry.Register(dsc); err != nil {
		return nil, fmt.Errorf("couldn't register ds8k collector: %s", err)
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{h.exporterMetricsRegistry, registry},
		promhttp.HandlerOpts{
			ErrorLog:      log.NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
			// MaxRequestsInFlight: h.maxRequests,
		},
	)
	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}
