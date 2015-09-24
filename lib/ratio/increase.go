package ratio

func Increase(xs []float64) float64 {
	return (xs[len(xs)-1] - xs[0]) / float64(len(xs))
}
