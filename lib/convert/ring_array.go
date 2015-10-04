package convert

import "container/ring"

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
