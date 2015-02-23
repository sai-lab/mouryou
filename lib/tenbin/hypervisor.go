package tenbin

import (
	"../math"
	"sync"
)

type hypervisor struct {
	Host string
	VMs  []virtualMachine
}

func (h hypervisor) operatingRatios() []float64 {
	var wg sync.WaitGroup
	ors := make([]float64, len(h.VMs))

	for i, vm := range h.VMs {
		wg.Add(1)
		go func(i int, vm virtualMachine) {
			ors[i] = vm.operatingRatio()
			wg.Done()
		}(i, vm)
	}

	wg.Wait()
	return ors
}

func (h hypervisor) avgor() float64 {
	ors := h.operatingRatios()
	return math.Average(ors)
}
