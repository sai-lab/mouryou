package calculate

func Average(xs []float64) float64 {
	total := float64(0)

	for _, x := range xs {
		total += x
	}

	return total / float64(len(xs))
}

func MovingAverage(xs []float64, n int) float64 {
	if len(xs) < n {
		return Average(xs)
	} else {
		return Average(xs[len(xs)-n:])
	}
}
