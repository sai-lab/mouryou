package models

import (
	"sync"
)

type ClusterStruct struct {
	LoadBalancer    LoadBalancerStruct `json:"load_balancer"`
	Hypervisors     []HypervisorStruct `json:"hypervisors"`
	VirtualMachines []VirtualMachineStruct
}

func (cluster *ClusterStruct) Initialize() {
	for _, hypervisor := range cluster.Hypervisors {
		hypervisor.Initialize()
		cluster.VirtualMachines = append(cluster.VirtualMachines, hypervisor.VirtualMachines...)
	}

	cluster.LoadBalancer.Initialize()
	for _, machine := range cluster.VirtualMachines {
		cluster.LoadBalancer.Add(machine.Host)
	}

	cluster.LoadBalancer.Active(cluster.VirtualMachines[0].Host)
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
