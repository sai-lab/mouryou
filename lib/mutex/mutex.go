package mutex

import "sync"

func Read(x *int, mutex *sync.RWMutex) int {
	mutex.RLock()
	defer mutex.RUnlock()

	return *x
}

func Write(x *int, mutex *sync.RWMutex, y int) {
	mutex.Lock()
	defer mutex.Unlock()

	*x = y
}

func ReadFloat(x *float64, mutex *sync.RWMutex) float64 {
	mutex.RLock()
	defer mutex.RUnlock()

	return *x
}

func WriteFloat(x *float64, mutex *sync.RWMutex, y float64) {
	mutex.Lock()
	defer mutex.Unlock()

	*x = y
}
