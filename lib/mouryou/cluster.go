package mouryou

import (
	"sync"
)

type cluster struct {
	Timeout int64
	Working int
	LB      loadBalancer
	HVs     []hypervisor
	VMs     []virtualMachine
}

func (c *cluster) init() {
	for _, hv := range c.HVs {
		hv.assignVMs()
		c.VMs = append(c.VMs, hv.VMs...)
	}
	c.Working = 1
}

func (c cluster) operatingRatios() []float64 {
	var wg sync.WaitGroup
	var m sync.Mutex
	ors := make([]float64, c.Working)

	for i := 0; i < c.Working; i++ {
		wg.Add(1)
		go func(i int, vm virtualMachine) {
			defer wg.Done()
			or := vm.operatingRatio()

			m.Lock()
			ors[i] = or
			m.Unlock()
		}(i, c.VMs[i])
	}
	wg.Wait()

	return ors
}
