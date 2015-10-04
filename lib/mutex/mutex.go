package mutex

import "sync"

func Read(x *int, mutex *sync.RWMutex) int {
	mutex.RLock()
	defer mutex.RUnlock()

	return *x
}

func Write(x *int, y int, mutex *sync.RWMutex) {
	mutex.Lock()
	defer mutex.Unlock()

	*x = y
}
