package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/tidwall/gjson"
	"github.ibm.com/ZaaS/ds8k-exporter/utils"
)

const (
	prefixVolume                  = "ds8k_volume_"
	totalVolumeCapacityName       = "capacity_total"
	allocatedVolumeCapacityName   = "capacity_allocated"
	volumeCapacityUsedPercentName = "capacity_used_percent"
	totalVolumeCapacityDesc       = "The total capacity of volume."
	allocatedVolumeCapacityDesc   = "The allocated capacity of volume."
	volumeCapacityUsedPercentDesc = "The volume capacity utilization."
)

var (
	totalVolumeCapacity       *prometheus.Desc
	allocatedVolumeCapacity   *prometheus.Desc
	volumeCapacityUsedPercent *prometheus.Desc
)

func init() {
	registerCollector("volume", defaultEnabled, NewVolumeCollector)
	labelnames := []string{"target", "volume", "pool"}
	totalVolumeCapacity = prometheus.NewDesc(prefixVolume+totalVolumeCapacityName, totalVolumeCapacityDesc, labelnames, nil)
	allocatedVolumeCapacity = prometheus.NewDesc(prefixVolume+allocatedVolumeCapacityName, allocatedVolumeCapacityDesc, labelnames, nil)
	volumeCapacityUsedPercent = prometheus.NewDesc(prefixVolume+volumeCapacityUsedPercentName, volumeCapacityUsedPercentDesc, labelnames, nil)
}

// poolCollector collects system metrics
type volumeCollector struct {
}

func NewVolumeCollector() (Collector, error) {
	return &volumeCollector{}, nil
}

//Describe describes the metrics
func (*volumeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- totalVolumeCapacity
	ch <- allocatedVolumeCapacity
	ch <- volumeCapacityUsedPercent
}

//Collect collects metrics from DS8k Restful API
func (c *volumeCollector) Collect(dClient utils.DS8kClient, ch chan<- prometheus.Metric) {
	log.Debugln("Entering volumes collector ...")
	reqPoolURL := "https://" + dClient.IpAddress + ":" + ds8KAPIPort + "/api/v1/pools"
	poolsResp, err := dClient.CallDS8kAPI(reqPoolURL)
	if err != nil {
		log.Errorln("Executing '/api/v1/pools' request failed: ", err)
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
	var poolIds []string
	for _, pool := range pools {
		poolIds = append(poolIds, pool.Get("id").String())
		poolID := pool.Get("id").String()
		requestVolume := "/api/v1/pools/" + poolID + "/volumes"
		reqVolumeURL := "https://" + dClient.IpAddress + ":" + ds8KAPIPort + requestVolume
		VolumesResp, err := dClient.CallDS8kAPI(reqVolumeURL)
		if err != nil {
			log.Errorln("Executing '/api/v1/pools/"+poolID+"/volumes' request failed: %s", err)
		}
		log.Debugln("result of '/api/v1/pools/"+poolID+"/volumes': ", poolsResp)
		// This is the sample output of /api/v1/pools/poolID/volumes
		// {
		// 	"counts": {
		// 		"data_counts": 19,
		// 		"total_counts": 19
		// 	},
		// 	"data": {
		// 		"volumes": [
		// 			{
		// 				"MTM": "2107-900",
		// 				"VOLSER": "",
		// 				"allocmethod": "rotateexts",
		// 				"cap": "53687091200",
		// 				"capalloc": "53687091200",
		// 				"datatype": "FB 512",
		// 				"easytier": "none",
		// 				"id": "0002",
		// 				"link": {
		// 					"href": "https:/10.23.1.10:8452/api/v1/volumes/0002",
		// 					"rel": "self"
		// 				},
		// 				"lss": {
		// 					"id": "00",
		// 					"link": {
		// 						"href": "https:/10.23.1.10:8452/api/v1/lss/00",
		// 						"rel": "self"
		// 					}
		// 				},
		// 				"name": "mgr_hm1_code",
		// 				"pool": {
		// 					"id": "P0",
		// 					"link": {
		// 						"href": "https:/10.23.1.10:8452/api/v1/pools/P0",
		// 						"rel": "self"
		// 					}
		// 				},
		// 				"real_cap": "53687091200",
		// 				"state": "normal",
		// 				"stgtype": "fb",
		// 				"tieralloc": [
		// 					{
		// 						"allocated": "53687091200",
		// 						"tier": "ENT"
		// 					}
		// 				],
		// 				"tp": "none",
		// 				"virtual_cap": "0"
		// 			}
		// 			]
		// 		},
		// 		"server": {
		// 			"code": "",
		// 			"message": "Operation done successfully.",
		// 			"status": "ok"
		// 		}
		// 	}

		volumesData := gjson.Get(VolumesResp, "data").String()
		volumes := gjson.Get(volumesData, "volumes").Array()
		for _, volume := range volumes {
			labelvalues := []string{dClient.IpAddress, volume.Get("name").String() + "_" + volume.Get("id").String(), pool.Get("name").String() + "_" + pool.Get("id").String()}
			ch <- prometheus.MustNewConstMetric(totalVolumeCapacity, prometheus.GaugeValue, volume.Get("cap").Float(), labelvalues...)
			ch <- prometheus.MustNewConstMetric(allocatedVolumeCapacity, prometheus.GaugeValue, volume.Get("capalloc").Float(), labelvalues...)
			ch <- prometheus.MustNewConstMetric(volumeCapacityUsedPercent, prometheus.GaugeValue, volume.Get("capalloc").Float()/volume.Get("cap").Float(), labelvalues...)
		}

	}
	log.Debugln("Leaving volumes collector.")
}
