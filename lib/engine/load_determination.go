package engine

import (
	"github.com/sai-lab/mouryou/lib/models"
)

func LoadDetermination(config *models.Config) {
	if config.UseOperatingRatio {
		go operatingRatioBase(config)
	}
	if config.UseThroughput {
		go throughputBase(config)
	}
}
