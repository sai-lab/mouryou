package engine

import (
	"sync"

	"github.com/sai-lab/mouryou/lib/monitor"
)

func StatusManager() {
	var mutex sync.RWMutex

	for status := range monitor.StateCh {
		name := status.Name
		mutex.Lock()
		for i, v := range monitor.States {
			if v.Name == name {
				monitor.States[i].Weight = status.Weight
				monitor.States[i].Info = status.Info
				break
			}
		}
		mutex.Unlock()
	}
}
