package tenbin

import (
	"../math"
	"log"
	"sync"
)

type cluster struct {
	Threshold float64
	LB        loadBalancer
	HVs       []hypervisor
}

func (c cluster) operatingRatios() []float64 {
	var wg sync.WaitGroup
	ors := make([]float64, len(c.HVs))

	for i, hv := range c.HVs {
		wg.Add(1)
		go func(i int, hv hypervisor) {
			ors[i] = hv.avgor()
			wg.Done()
		}(i, hv)
	}

	wg.Wait()
	return ors
}

func (c cluster) avgor() float64 {
	ors := c.operatingRatios()
	return math.Average(ors)
}

func (c cluster) Log() {
	log.Printf("%.5f\n", c.avgor())
}
