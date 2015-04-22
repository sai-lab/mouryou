package mouryou

import (
	"container/ring"
	"fmt"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rtoa(r *ring.Ring) []float64 {
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
