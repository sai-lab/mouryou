package mouryou

import (
	"sync"
)

var working int = 1
var workM sync.RWMutex

var operating int = 0
var operateM sync.RWMutex

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

func readOperating() int {
	operateM.RLock()
	defer operateM.RUnlock()
	return operating
}

func writeOperating(o int) {
	operateM.Lock()
	operating = o
	defer operateM.Unlock()
}
