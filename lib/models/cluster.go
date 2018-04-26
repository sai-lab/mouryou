package models

import (
	"sync"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type Cluster struct {
	LoadBalancer    LoadBalancer              `json:"load_balancer"`
	Vendors         []VendorStruct            `json:"vendors"`
	Hypervisors     []HypervisorStruct        `json:"hypervisors"`
	VirtualMachines map[string]VirtualMachine `json:"virtual_machines"`
}

func (cluster *Cluster) Initialize(config *Config) {
	cluster.LoadBalancer.Initialize()
	for _, vendor := range cluster.Vendors {
		vendor.Initialize()
		cluster.VirtualMachines = vendor.VirtualMachines
	}

	for _, machine := range cluster.VirtualMachines {
		cluster.LoadBalancer.Add(machine.Name)
		logger.PrintPlace("machine Name:" + machine.Name)
		if config.ContainID(machine.Id) {
			logger.PrintPlace("machine Name:" + machine.Name)
			continue
		}
		cluster.LoadBalancer.Inactive(machine.Name)
	}
}

func (cluster *Cluster) valueChack() error {
	return nil
}

// ServerStatuses は稼働中のサーバ配列btを受け取り、btの負荷状況を返します。
func (cluster Cluster) ServerStatuses(bt []string) []apache.ServerStatus {
	var group sync.WaitGroup
	var mutex sync.Mutex
	statuses := make([]apache.ServerStatus, len(bt))

	for i, v := range bt {
		group.Add(1)
		go func(i int, v string, machines map[string]VirtualMachine) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()
			for _, machine := range machines {
				if machine.Name == v {
					statuses[i] = machine.ServerState()
				}
			}
		}(i, v, cluster.VirtualMachines)
	}
	group.Wait()

	return statuses
}
