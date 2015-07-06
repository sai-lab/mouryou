package mouryou

import (
	"container/ring"
	"fmt"
	"os"
	"sync"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func readWithMutex(x *int, mutex *sync.RWMutex) int {
	mutex.RLock()
	defer mutex.RUnlock()

	return *x
}

func writeWithMutex(x *int, y int, mutex *sync.RWMutex) {
	mutex.Lock()
	defer mutex.Unlock()

	*x = y
}

func RingToArray(r *ring.Ring) []float64 {
	arr := make([]float64, r.Len())
	i := 0

	r.Do(func(v interface{}) {
		if v != nil {
			arr[i] = v.(float64)
			i++
		}
	})

	return arr[:i]
}

func FloatsToStrings(xs []float64) []string {
	arr := make([]string, len(xs))

	for i, x := range xs {
		arr[i] = fmt.Sprintf("%.5f", x)
	}

	return arr
}
