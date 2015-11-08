package ratio

import "github.com/sai-lab/mouryou/lib/average"

func Increase(xs []float64, n int) float64 {
	return (average.MovingAverage(xs, n) - xs[0]) / float64(len(xs))
}
