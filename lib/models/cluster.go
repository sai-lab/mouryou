package models

import (
	"sync"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type ClusterStruct struct {
	LoadBalancer    LoadBalancerStruct              `json:"load_balancer"`
	Vendors         []VendorStruct                  `json:"vendors"`
	Hypervisors     []HypervisorStruct              `json:"hypervisors"`
	VirtualMachines map[string]VirtualMachineStruct `json:"virtual_machines"`
}

func (cluster *ClusterStruct) Initialize() {
	cluster.LoadBalancer.Initialize()
	logger.PrintPlace("Cluster Initialize")
	for _, vendor := range cluster.Vendors {
		vendor.Initialize()
		cluster.VirtualMachines = vendor.VirtualMachines
	}

	for _, machine := range cluster.VirtualMachines {
		cluster.LoadBalancer.Add(machine.Name)
		if machine.Id == 1 {
			continue
		}
		cluster.LoadBalancer.Inactive(machine.Name)
	}
}

func (cluster ClusterStruct) ServerStates(n int) []apache.ServerStat {
	var group sync.WaitGroup
	var mutex sync.Mutex
	states := make([]apache.ServerStat, n)

	for i := 0; i < n; i++ {
		group.Add(1)
		go func(i int, machines map[string]VirtualMachineStruct) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()
			for _, machine := range machines {
				if machine.Id == i+1 {
					states[i] = machine.ServerState()
				}
			}
		}(i, cluster.VirtualMachines)
	}
	group.Wait()

	return states
}
