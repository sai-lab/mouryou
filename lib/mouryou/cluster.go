package mouryou

import (
	"sync"
)

type cluster struct {
	LB  loadBalancer
	HVs []hypervisor
	VMs []virtualMachine
}

func (c *cluster) init() {
	for _, hv := range c.HVs {
		hv.init()
		c.VMs = append(c.VMs, hv.VMs...)
	}

	c.LB.init()
	for _, vm := range c.VMs {
		c.LB.add(vm.Host)
	}

	c.LB.active(c.VMs[0].Host)
}

func (c cluster) operatingRatios(working int) []float64 {
	var wg sync.WaitGroup
	var m sync.Mutex
	ors := make([]float64, working)

	for i := 0; i < working; i++ {
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
