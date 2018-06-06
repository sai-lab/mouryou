package monitor

import (
	"strconv"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

// WeightMonitor は重みを計測する関数です。
func WeightMonitor(config *models.Config) {
	for _, state := range States {
		if state.Name != "" {
			if config.DevelopLogLevel >= 5 {
				logger.PrintPlace("state Name: " + state.Name + ", state weight: " + strconv.Itoa(state.Weight))
			}
		}
	}
}
