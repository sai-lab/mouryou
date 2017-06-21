package functions

import (
	"strconv"
	"sync"

	"github.com/sai-lab/mouryou/lib/logger"
)

func StatusManager() {
	var status StatusStruct
	var mutex sync.RWMutex

	for status = range statusCh {
		sts := []string{"st", status.Name, strconv.FormatInt(int64(status.Weight), 10), status.Info}
		logger.Print(sts)
		logger.Write(sts)
		name := status.Name
		mutex.Lock()
		for i, v := range states {
			if v.Name == name {
				states[i].Weight = status.Weight
				states[i].Info = status.Info
				break
			}
		}
		mutex.Unlock()
	}
}
