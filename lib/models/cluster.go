package models

import (
	"sync"

	"github.com/sai-lab/mouryou/lib/logger"
)

type ClusterStruct struct {
	LoadBalancer    LoadBalancerStruct     `json:"load_balancer"`
	Vendors         []VendorStruct         `json:"vendors"`
	Hypervisors     []HypervisorStruct     `json:"hypervisors"`
	VirtualMachines []VirtualMachineStruct `json:"-"`
}

func (cluster *ClusterStruct) Initialize() {
	cluster.LoadBalancer.Initialize()
	logger.PrintPlace("Cluster Initialize")
	for _, vendor := range cluster.Vendors {
		logger.PrintPlace("range cluster.Vendors")
		vendor.Initialize()
		cluster.VirtualMachines = append(cluster.VirtualMachines, vendor.VirtualMachines...)
	}

	for _, machine := range cluster.VirtualMachines {
		cluster.LoadBalancer.Add(machine.Host)
		cluster.LoadBalancer.Inactive(machine.Name)
	}

	cluster.LoadBalancer.Active(cluster.VirtualMachines[0].Name)
}

func (cluster ClusterStruct) OperatingRatios(n int) []float64 {
	var group sync.WaitGroup
	var mutex sync.Mutex
	ratios := make([]float64, n)

	for i := 0; i < n; i++ {
		group.Add(1)
		go func(i int, machine *VirtualMachineStruct) {
			defer group.Done()

			mutex.Lock()
			defer mutex.Unlock()
			ratios[i] = machine.OperatingRatio()
		}(i, &cluster.VirtualMachines[i])
	}
	group.Wait()

	return ratios
}
