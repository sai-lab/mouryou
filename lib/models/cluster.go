package models

import (
	"sync"

	"github.com/sai-lab/mouryou/lib/apache"
)

type Cluster struct {
	LoadBalancer    LoadBalancerStruct              `json:"load_balancer"`
	Vendors         []VendorStruct                  `json:"vendors"`
	Hypervisors     []HypervisorStruct              `json:"hypervisors"`
	VirtualMachines map[string]VirtualMachineStruct `json:"virtual_machines"`
}

func (cluster *Cluster) Initialize() {
	cluster.LoadBalancer.Initialize()
	for _, vendor := range cluster.Vendors {
		vendor.Initialize()
		cluster.VirtualMachines = vendor.VirtualMachines
	}

	for _, machine := range cluster.VirtualMachines {
		cluster.LoadBalancer.Add(machine.Name)
		// if machine.Id == 1 || machine.Id == 2 || machine.Id == 3 {
		// 	continue
		// }
		if machine.Id == 1 {
			continue
		}
		cluster.LoadBalancer.Inactive(machine.Name)
	}
}

func (cluster Cluster) ServerStates(bt []string) []apache.ServerStat {
	var group sync.WaitGroup
	var mutex sync.Mutex
	states := make([]apache.ServerStat, len(bt))

	for i, v := range bt {
		group.Add(1)
		go func(i int, v string, machines map[string]VirtualMachineStruct) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()
			for _, machine := range machines {
				if machine.Name == v {
					states[i] = machine.ServerState()
				}
			}
		}(i, v, cluster.VirtualMachines)
	}
	group.Wait()

	return states
}
