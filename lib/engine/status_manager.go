package engine

import (
	"strconv"
	"sync"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/monitor"
)

func StatusManager() {
	var mutex sync.RWMutex

	for status := range monitor.StatusCh {
		sts := []string{"st", status.Name, strconv.FormatInt(int64(status.Weight), 10), status.Info}
		logger.Print(sts)
		logger.Write(sts)
		name := status.Name
		mutex.Lock()
		for i, v := range monitor.Statuses {
			if v.Name == name {
				monitor.Statuses[i].Weight = status.Weight
				monitor.Statuses[i].Info = status.Info
				break
			}
		}
		mutex.Unlock()
	}
}
