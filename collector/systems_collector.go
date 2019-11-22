package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/tidwall/gjson"
	"github.ibm.com/ZaaS/ds8k-exporter/utils"
)

const (
	prefixSys                     = "ds8k_system_"
	totalSystemCapacityName       = "capacity_total"
	availableSystemCapacityName   = "capacity_available"
	allocatedSystemCapacityName   = "capacity_allocated"
	systemCapacityUsedPercentName = "capacity_used_percent"
	rawSystemCapacityName         = "capacity_raw"
	totalSystemCapacityDesc       = "The total capacity of system"
	availableSystemCapacityDesc   = "The avaliable capacity of system"
	allocatedSystemCapacityDesc   = "The allocated capacity of system"
	systemCapacityUsedPercentDesc = "The system capacity utilization."
	rawSystemCapacityDesc         = "The raw capacity of system"
)

var (
	totalSystemCapacity       *prometheus.Desc
	availableSystemCapacity   *prometheus.Desc
	allocatedSystemCapacity   *prometheus.Desc
	systemCapacityUsedPercent *prometheus.Desc
	rawSystemCapacity         *prometheus.Desc
)

func init() {
	registerCollector("system", defaultEnabled, NewSystemCollector)
	labelnames := []string{"resource", "target"}
	totalSystemCapacity = prometheus.NewDesc(prefixSys+totalSystemCapacityName, totalSystemCapacityDesc, labelnames, nil)
	availableSystemCapacity = prometheus.NewDesc(prefixSys+availableSystemCapacityName, availableSystemCapacityDesc, labelnames, nil)
	allocatedSystemCapacity = prometheus.NewDesc(prefixSys+allocatedSystemCapacityName, allocatedSystemCapacityDesc, labelnames, nil)
	systemCapacityUsedPercent = prometheus.NewDesc(prefixSys+systemCapacityUsedPercentName, systemCapacityUsedPercentDesc, labelnames, nil)
	rawSystemCapacity = prometheus.NewDesc(prefixSys+rawSystemCapacityName, rawSystemCapacityDesc, labelnames, nil)
}

// poolCollector collects system metrics
type systemCollector struct {
}

func NewSystemCollector() (Collector, error) {
	return &systemCollector{}, nil
}

//Describe describes the metrics
func (*systemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalSystemCapacity
	ch <- availableSystemCapacity
	ch <- allocatedSystemCapacity
	ch <- systemCapacityUsedPercent
	ch <- rawSystemCapacity
}

//Collect collects metrics from DS8k Restful API
func (c *systemCollector) Collect(dClient utils.DS8kClient, ch chan<- prometheus.Metric) {
	log.Debugln("Entering systems collector ...")
	reqSystemURL := "https://" + dClient.IpAddress + ":" + ds8KAPIPort + "/api/v1/systems"
	systemsResp, err := dClient.CallDS8kAPI(reqSystemURL)
	if err != nil {
		log.Errorln("Executing '/api/v1/systems' failed: ", err)
	}
	log.Debugln("Response of '/api/v1/systems': ", systemsResp)
	// This is the sample output of /api/v1/systems
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
		labelvalues := []string{system.Get("name").String(), dClient.IpAddress}
		ch <- prometheus.MustNewConstMetric(totalSystemCapacity, prometheus.GaugeValue, system.Get("cap").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(availableSystemCapacity, prometheus.GaugeValue, system.Get("capavail").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(allocatedSystemCapacity, prometheus.GaugeValue, system.Get("capalloc").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(systemCapacityUsedPercent, prometheus.GaugeValue, system.Get("capalloc").Float()/system.Get("cap").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(rawSystemCapacity, prometheus.GaugeValue, system.Get("capraw").Float(), labelvalues...)
	}
	log.Debugln("Leaving systems collector.")
}
