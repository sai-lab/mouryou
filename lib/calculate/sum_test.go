package calculate

import (
	"fmt"
	"testing"
)

func TestSum(t *testing.T) {
	floats := []float64{1.1, 2.2, 3.3}
	if float64(6.6) != Sum(floats) {
		t.Fatal("failed test Sum returns " + fmt.Sprint(Sum(floats)))
	}
}
