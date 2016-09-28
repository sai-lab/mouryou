package ratio

import "github.com/sai-lab/mouryou/lib/calculate"

func Increase(xs []float64, n int) float64 {
	if len(xs) == 1 {
		return calculate.MovingAverage(xs, n) - xs[0]
	}
	return (calculate.MovingAverage(xs, n) - xs[0]) / float64(len(xs)-1)
}
