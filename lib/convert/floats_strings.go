package convert

import "fmt"

func FloatsToStrings(xs []float64) []string {
	arr := make([]string, len(xs))

	for i, x := range xs {
		arr[i] = fmt.Sprintf("%.5f", x)
	}

	return arr
}
