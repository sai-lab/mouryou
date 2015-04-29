package mouryou

import (
	"sync"
)

var working int = 1
var workM sync.RWMutex

func readWorking() int {
	workM.RLock()
	defer workM.RUnlock()
	return working
}

func writeWorking(w int) {
	workM.Lock()
	working = w
	defer workM.Unlock()
}

var operating bool = false
var operateM sync.RWMutex

func readOperating() bool {
	operateM.RLock()
	defer operateM.RUnlock()
	return operating
}

func writeOperating(o bool) {
	operateM.Lock()
	operating = o
	defer operateM.Unlock()
}
