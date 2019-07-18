package collector

import (
	"regexp"
	"strings"
	"time"

	"github.com/ds8k-exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/tidwall/gjson"
)

const (
	prefixPerformance = "ds8k_performance_"
	readIOName        = "read"
	writeIOName       = "write"
	totalIOName       = "total"
	readIODesc        = "The average number of I/O operations that are transferred per second for read operations to Systems during the sample period."
	writeIODesc       = "The average number of I/O operations that are transferred per second for write operations to Systems during the sample period."
	totalIODesc       = "The average number of I/O operations that are transferred per second for read and write operations to Systems during the sample period."
)

var (
	read  *prometheus.Desc
	write *prometheus.Desc
	total *prometheus.Desc
)

func init() {
	registerCollector("performance", defaultEnabled, NewPerformanceCollector)
	labelnames := []string{"resource", "target"}
	read = prometheus.NewDesc(prefixPerformance+readIOName, readIODesc, labelnames, nil)
	write = prometheus.NewDesc(prefixPerformance+writeIOName, writeIODesc, labelnames, nil)
	total = prometheus.NewDesc(prefixPerformance+totalIOName, totalIODesc, labelnames, nil)
}

// poolCollector collects system metrics
type performanceCollector struct {
}

func NewPerformanceCollector() (Collector, error) {
	return &performanceCollector{}, nil
}

//Describe describes the metrics
func (*performanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- read
	ch <- write
	ch <- total
}

//Collect collects metrics from DS8k Restful API
func (c *performanceCollector) Collect(dClient utils.DS8kClient, ch chan<- prometheus.Metric) {
	log.Debugln("Entering performance collector ...")
	reqSystemURL := "https://" + dClient.IpAddress + ":" + ds8KAPIPort + "/api/v1/systems"
	systemsResp, err := dClient.CallDS8kAPI(reqSystemURL)
	if err != nil {
		log.Errorf("Executing /api/v1/systems request failed: %s", err)
	}
	log.Debugln("Response of '/api/v1/systems': ", systemsResp)
	//This is the sample output of /api/v1/systems call
	// {
	// 	"counts": {
	// 		"data_counts": 1,
	// 		"total_counts": 1
	// 	},
	// 	"data": {
	// 		"systems": [
	// 			{
	// 				"MTM": "2831-984",
	// 				"bundle": "88.33.41.0",
	// 				"cap": "43512313675776",
	// 				"capalloc": "31521838727168",
	// 				"capavail": "11906723086336",
	// 				"capraw": "87241523200000",
	// 				"id": "2107-75DXA41",
	// 				"name": "IBM.2107-75DXA40",
	// 				"release": "8.3.3",
	// 				"sn": "75DXA41",
	// 				"state": "online",
	// 				"wwnn": "5005076306FFD65A"
	// 			}
	// 		]
	// 	},
	// 	"server": {
	// 		"code": "",
	// 		"message": "Operation done successfully.",
	// 		"status": "ok"
	// 	}
	// }

	systemsData := gjson.Get(systemsResp, "data").String()
	systems := gjson.Get(systemsData, "systems").Array()
	for _, system := range systems {
		serial_number := system.Get("sn").String()
		location, err := time.LoadLocation(dClient.Location)
		//Examples of dClient.Location: America/New_York ; America/Los_Angeles
		if err != nil {
			log.Errorln("Loading location of device failed: ", err)
		}

		deviceTime := time.Now().In(location) //Get ds8k's location time.  Example: 2019-07-09 23:20:47.890562 -0400 EDT
		log.Debugln(" ds8k's local time is ", deviceTime)
		timeZone := regexp.MustCompile(`\+|\-\d{4}`).FindString(deviceTime.String()) // Get timezone from devicetime. Example: -0400
		duration, _ := time.ParseDuration("-1m")                                     //Roll back 1 minute
		beforeTime := deviceTime.Add(duration)
		afterTime := beforeTime.Add(duration)
		beforeTimeFormat := strings.Replace(regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`).FindString(beforeTime.String()), " ", "T", -1) // Get time form devicetime. Example: 2019-07-09 23:20:47
		afterTimeFormat := strings.Replace(regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`).FindString(afterTime.String()), " ", "T", -1)
		reqPerformanceURL := "https://" + dClient.IpAddress + ":8452/api/v1/systems/" + serial_number + "/performance?after=" + afterTimeFormat + timeZone + "&before=" + beforeTimeFormat + timeZone
		performanceInfo, err := dClient.CallDS8kAPI(reqPerformanceURL)
		if err != nil {
			log.Errorln("Executing '/api/v1/systems/"+serial_number+"/performance?after="+afterTimeFormat+timeZone+"&before="+beforeTimeFormat+timeZone+"' request failed: ", err)
		}
		log.Debugln("Response of '/api/v1/systems/"+serial_number+"/performance?after="+afterTimeFormat+timeZone+"&before="+beforeTimeFormat+timeZone+"' : ", performanceInfo)
		// This is the sample output of /api/v1/systems/performances?after=afterTime&before=beforeTime call
		// {
		// 	"counts": {
		// 		"data_counts": 1,
		// 		"total_counts": 1
		// 	},
		// 	"data": {
		// 		"performance": [
		// 			{
		// 				"IOPS": {
		// 					"read": "4.38",
		// 					"total": "457",
		// 					"write": "452.62"
		// 				},
		// 				"performancesampletime": "2019-05-20T01:44:42-0400",
		// 				"responseTime": {
		// 					"average": "0.44",
		// 					"read": "0",
		// 					"write": "0.44"
		// 				}
		// 			}
		// 		]
		// 	},
		// 	"server": {
		// 		"code": "",
		// 		"message": "Operation done successfully.",
		// 		"status": "ok"
		// 	}
		// }

		performanceData := gjson.Get(performanceInfo, "data").String()
		performances := gjson.Get(performanceData, "performance").Array()
		labelvalues := []string{system.Get("name").String(), dClient.IpAddress}
		if len(performances) > 0 {
			IOPS := performances[0].Get("IOPS")
			ch <- prometheus.MustNewConstMetric(read, prometheus.GaugeValue, IOPS.Get("read").Float(), labelvalues...)
			ch <- prometheus.MustNewConstMetric(write, prometheus.GaugeValue, IOPS.Get("write").Float(), labelvalues...)
			ch <- prometheus.MustNewConstMetric(total, prometheus.GaugeValue, IOPS.Get("total").Float(), labelvalues...)
		} else {
			log.Errorln("Metric of performance is null")
		}

	}
	log.Debugln("Leaving performance collector.")
}
