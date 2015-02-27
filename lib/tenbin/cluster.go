package tenbin

import (
	"../math"
	"log"
	"sync"
)

type cluster struct {
	Threshold float64
	Timeout   int64
	LB        loadBalancer
	HVs       []hypervisor
	VMs       []virtualMachine
}

func (c *cluster) InitVMs() {
	for _, hv := range c.HVs {
		c.VMs = append(c.VMs, hv.VMs...)
	}
}

func (c cluster) operatingRatios() []float64 {
	var wg sync.WaitGroup
	ors := make([]float64, len(c.VMs))

	for i, vm := range c.VMs {
		wg.Add(1)
		go func(i int, vm virtualMachine) {
			ors[i] = vm.operatingRatio()
			wg.Done()
		}(i, vm)
	}

	wg.Wait()
	return ors
}

func (c cluster) avgor() float64 {
	ors := c.operatingRatios()
	log.Printf("%+v\n", ors)
	return math.Average(ors)
}

func (c cluster) Log() {
	log.Printf("%.5f\n", c.avgor())
}
