package mouryou

import (
	"sync"
)

var working int = 1
var rwm sync.RWMutex

func readWorking() int {
	rwm.RLock()
	defer rwm.RUnlock()
	return working
}

func writeWorking(w int) {
	rwm.Lock()
	working = w
	defer rwm.Unlock()
}
