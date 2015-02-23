package tenbin

import (
	"../math"
	"fmt"
	"log"
	"strings"
	"sync"
)

type Hypervisor struct {
	vms []virtualMachine
}

func (h *Hypervisor) AddVM(name string, ipAddress string) {
	var vm virtualMachine = virtualMachine{name, ipAddress}
	h.vms = append(h.vms, vm)
}

func (h Hypervisor) operatingRatios() []float64 {
	var wg sync.WaitGroup
	ors := make([]float64, len(h.vms))

	for i, vm := range h.vms {
		wg.Add(1)
		go func(i int, vm virtualMachine) {
			ors[i] = vm.operatingRatio()
			wg.Done()
		}(i, vm)
	}

	wg.Wait()
	return ors
}

func (h Hypervisor) AVGOR() float64 {
	ors := h.operatingRatios()
	return math.Average(ors)
}

func (h Hypervisor) PrintLoads() {
	ors := h.operatingRatios()
	loads := make([]string, len(h.vms))

	for i := range ors {
		loads[i] = fmt.Sprintf("%.5f", ors[i])
	}

	log.Println(strings.Join(loads, ", "))
}
