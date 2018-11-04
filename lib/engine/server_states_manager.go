package engine

import (
	"sync"

	"time"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/monitor"
)

func ServerStatesManager() {
	var mutex sync.RWMutex

	for state := range monitor.StateCh {
		mutex.Lock()
		// TODO 確認
		zeroTime, err := time.Parse("January 2, year 2006, 15:04:05 MST", "January 1, year 1, 00:00:00 UTC")
		if err != nil {
			logger.Error(logger.Place(), err)
		}
		err = monitor.UpdateServerStates(state.Name, state.Weight, state.Info, zeroTime)
		mutex.Unlock()
		if err != nil {
			logger.Error(logger.Place(), err)
		}
	}
}
