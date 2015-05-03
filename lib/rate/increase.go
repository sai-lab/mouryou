package rate

import (
	"math"
)

func Increase(xs []float64) int {
	r := xs[len(xs)-1] / xs[0]

	if r <= 0 {
		return 1
	} else {
		return int(math.Ceil(r))
	}
}
