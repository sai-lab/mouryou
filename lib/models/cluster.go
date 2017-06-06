package models

import (
	"sync"

	"github.com/sai-lab/mouryou/lib/apache"
)

type ClusterStruct struct {
	LoadBalancer    LoadBalancerStruct              `json:"load_balancer"`
	Vendors         []VendorStruct                  `json:"vendors"`
	Hypervisors     []HypervisorStruct              `json:"hypervisors"`
	VirtualMachines map[string]VirtualMachineStruct `json:"virtual_machines"`
}

func (cluster *ClusterStruct) Initialize() {
	cluster.LoadBalancer.Initialize()
	for _, vendor := range cluster.Vendors {
		vendor.Initialize()
		cluster.VirtualMachines = vendor.VirtualMachines
	}

	for _, machine := range cluster.VirtualMachines {
		cluster.LoadBalancer.Add(machine.Name)
		if machine.Id == 6 {
			continue
		}
		cluster.LoadBalancer.Inactive(machine.Name)
	}
	// cluster.LoadBalancer.Inactive("sai-server-2")
	// cluster.LoadBalancer.Inactive("sai-server-4")
	// cluster.LoadBalancer.Inactive("sai-server-7")
	// cluster.LoadBalancer.Inactive("sai-server-8")
	// cluster.LoadBalancer.Inactive("sai-server-9")
	// cluster.LoadBalancer.Inactive("sai-server-10")
	// cluster.LoadBalancer.Inactive("sai-server-11")
	// cluster.LoadBalancer.Inactive("sai-server-12")
}

func (cluster ClusterStruct) ServerStates(bt []string) []apache.ServerStat {
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
