package convert

import (
	"strconv"
)

func MapToArray(m map[string]int) []string {
	a := []string{}
	for k, v := range m {
		a = append(a, k+" "+strconv.Itoa(v))
	}
	return a
}
