package functions

import (
	"fmt"
	"sync"
	//"github.com/sai-lab/mouryou/lib/logger"
)

func StatusManager() {
	var status StatusStruct
	var mutex sync.RWMutex

	for status = range statusCh {
		mutex.Lock()
		defer mutex.Unlock()
		fmt.Println("StatusManager : statusCh get, Name: " + status.Name + ", Weight: " + fmt.Sprint(status.Weight) + ", Info: " + status.Info)
		name := status.Name
		for i, v := range states {
			if v.Name == name {
				states[i].Weight = status.Weight
				states[i].Info = status.Info
			}
		}
	}

}
