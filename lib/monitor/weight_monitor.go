package monitor

import (
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

// WeightMonitor は重みを計測する関数です。
func WeightMonitor(config *models.Config) {
	for _ = range StateCh {
		length := len(config.Cluster.VirtualMachines)
		weights := map[string]int{}
		weights["weights"] = -1
		for i, state := range States {
			weights[state.Name] = state.Weight
			if i == length-1 {
				break
			}
		}
		ar := convert.MapToArray(weights)
		logger.Write(ar)
		logger.Print(ar)
	}
}
