package collector

import (
	"github.com/ds8k-exporter/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/tidwall/gjson"
)

const prefixPool = "ds8k_pool_"
const totalPoolCapacityName = "capacity_total"
const availablePoolCapacityName = "capacity_available"
const allocatedPoolCapacityName = "capacity_allocated"
const poolCapacityUsedPercentName = "capacity_used_percent"
const totalPoolCapacityDesc = "The total capacity of pool"
const availablePoolCapacityDesc = "The avaliable capacity of pool"
const allocatedPoolCapacityDesc = "The allocated capacity of pool"
const poolCapacityUsedPercentDesc = "The pool capacity utilization."

var (
	totalPoolCapacity       *prometheus.Desc
	availablePoolCapacity   *prometheus.Desc
	allocatedPoolCapacity   *prometheus.Desc
	poolCapacityUsedPercent *prometheus.Desc
)

func init() {
	registerCollector("pool", defaultEnabled, NewPoolCollector)
	labelnames := []string{"target", "pool", "node"}
	totalPoolCapacity = prometheus.NewDesc(prefixPool+totalPoolCapacityName, totalPoolCapacityDesc, labelnames, nil)
	availablePoolCapacity = prometheus.NewDesc(prefixPool+availablePoolCapacityName, availablePoolCapacityDesc, labelnames, nil)
	allocatedPoolCapacity = prometheus.NewDesc(prefixPool+allocatedPoolCapacityName, allocatedPoolCapacityDesc, labelnames, nil)
	poolCapacityUsedPercent = prometheus.NewDesc(prefixPool+poolCapacityUsedPercentName, poolCapacityUsedPercentDesc, labelnames, nil)
}

// poolCollector collects system metrics
type poolCollector struct {
}

func NewPoolCollector() (Collector, error) {
	return &poolCollector{}, nil
}

//Describe describes the metrics
func (*poolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalPoolCapacity
	ch <- availablePoolCapacity
	ch <- allocatedPoolCapacity
	ch <- poolCapacityUsedPercent
}

//Collect collects metrics from DS8k Restful API
func (c *poolCollector) Collect(dClient utils.DS8kClient, ch chan<- prometheus.Metric) {
	log.Debugln("Entering pools collector ...")
	reqPoolURL := "https://" + dClient.IpAddress + ":" + ds8KAPIPort + "/api/v1/pools"
	poolsResp, err := dClient.CallDS8kAPI(reqPoolURL)
	if err != nil {
		log.Errorln("Executing '/api/v1/pools' failed: ", err)
	}
	log.Debugln("Response of '/api/v1/pools': ", poolsResp)
	// This is a sample output of /api/v1/polls call
	// {
	// 	"counts": {
	// 		"data_counts": 6,
	// 		"total_counts": 6
	// 	},
	// 	"data": {
	// 		"pools": [
	// 			{
	// 				"cap": "5541581553664",
	// 				"capalloc": "1248761741312",
	// 				"capavail": "4273492459520",
	// 				"easytier": "none",
	// 				"eserep": {},
	// 				"extent_size": "1GiB",
	// 				"id": "P0",
	// 				"link": {
	// 					"href": "https:/10.23.1.10:8452/api/v1/pools/P0",
	// 					"rel": "self"
	// 				},
	// 				"name": "Prod_code",
	// 				"node": "0",
	// 				"overprovisioned": "0.2",
	// 				"real_capacity_allocated_on_ese": "0",
	// 				"stgtype": "fb",
	// 				"threshold": "15",
	// 				"tieralloc": [
	// 					{
	// 						"allocated": "1248761741312",
	// 						"assigned": "0",
	// 						"cap": "5541581553664",
	// 						"tier": "ENT"
	// 					}
	// 				],
	// 				"tserep": {},
	// 				"virtual_capacity_allocated_on_ese": "0",
	// 				"volumes": {
	// 					"link": {
	// 						"href": "https:/10.23.1.10:8452/api/v1/pools/P0/volumes",
	// 						"rel": "self"
	// 					}
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

	poolsData := gjson.Get(poolsResp, "data").String()
	pools := gjson.Get(poolsData, "pools").Array()
	for _, pool := range pools {
		labelvalues := []string{dClient.IpAddress, pool.Get("name").String() + "_" + pool.Get("id").String(), pool.Get("node").String()}
		ch <- prometheus.MustNewConstMetric(totalPoolCapacity, prometheus.GaugeValue, pool.Get("cap").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(availablePoolCapacity, prometheus.GaugeValue, pool.Get("capavail").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(allocatedPoolCapacity, prometheus.GaugeValue, pool.Get("capalloc").Float(), labelvalues...)
		ch <- prometheus.MustNewConstMetric(poolCapacityUsedPercent, prometheus.GaugeValue, pool.Get("capalloc").Float()/pool.Get("cap").Float(), labelvalues...)
	}
	log.Debugln("Leaving pools collector.")
}
