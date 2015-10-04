package models

import (
	libvirt "github.com/alexzorin/libvirt-go"
)

type HypervisorStruct struct {
	Host            string                 `json:"host"`
	VirtualMachines []VirtualMachineStruct `json:"virtual_machines"`
}

func (hypervisor *HypervisorStruct) Initialize() {
	for i := range hypervisor.VirtualMachines {
		hypervisor.VirtualMachines[i].Hypervisor = hypervisor

		if i != 0 {
			hypervisor.VirtualMachines[i].Shutdown(0, nil)
		}
	}
}

func (hypervisor HypervisorStruct) Connect() (libvirt.VirConnection, error) {
	connection, err := libvirt.NewVirConnection("qemu+tcp://" + hypervisor.Host + "/system")
	return connection, err
}
