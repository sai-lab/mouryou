package convert

import "fmt"

func FloatsToStrings(xs []float64, s string) []string {
	arr := make([]string, len(xs)+1)
	arr[0] = s
	for i, x := range xs {
		arr[i+1] = fmt.Sprintf("%.5f", x)
	}

	return arr
}

func FloatsToStringsSimple(xs []float64) []string {
	arr := make([]string, len(xs))

	for i, x := range xs {
		arr[i] = fmt.Sprintf("%.5f", x)
	}

	return arr
}
