package monitor

import (
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

// WeightMonitor は重みを計測する関数です。
func WeightMonitor(config *models.Config) {
	for _ = range LoadORCh {
		length := len(config.Cluster.VirtualMachines)
		weights := map[string]int{}
		for i, state := range ServerStates {
			if state.Info != "booted up" {
				continue
			}
			weights[state.Name] = state.Weight
			if i == length-1 {
				break
			}
		}
		tags := []string{"operation:weight"}
		fields := convert.MapToArray(weights)
		logger.Record(tags, fields)
		databases.WriteValues(config.InfluxDBConnection, config, tags, fields)
	}
}
