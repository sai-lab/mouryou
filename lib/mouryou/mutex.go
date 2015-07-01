package mouryou

import (
	"sync"
)

var working int = 1
var workMutex sync.RWMutex

var operating int = 0
var operateMutex sync.RWMutex

func readWorking() int {
	workMutex.RLock()
	defer workMutex.RUnlock()
	return working
}

func writeWorking(x int) {
	workMutex.Lock()
	defer workMutex.Unlock()
	working = x
}

func readOperating() int {
	operateMutex.RLock()
	defer operateMutex.RUnlock()
	return operating
}

func writeOperating(x int) {
	operateMutex.Lock()
	defer operateMutex.Unlock()
	operating = x
}
