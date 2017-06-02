package functions

import (
	"fmt"
	"sync"
)

func StatusManager() {
	var status StatusStruct
	var mutex sync.RWMutex

	for status = range statusCh {
		fmt.Println("StatusManager : statusCh get, Name: " + status.Name + ", Weight: " + fmt.Sprint(status.Weight) + ", Info: " + status.Info)
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
