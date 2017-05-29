package models

//import libvirt "github.com/rgbkrk/libvirt-go"
import (
	"fmt"

	"github.com/sai-lab/mouryou/lib/logger"
)

type HypervisorStruct struct {
	Name            string                 `json:"name"`
	Host            string                 `json:"host"`
	VirtualMachines []VirtualMachineStruct `json:"virtual_machines"`
}

func (hypervisor *HypervisorStruct) Initialize() {
	logger.PrintPlace("Vendor Initialize")
	for i := range hypervisor.VirtualMachines {
		hypervisor.VirtualMachines[i].Hypervisor = hypervisor

		logger.PrintPlace(fmt.Sprint(i))
		if i != 0 {
			hypervisor.VirtualMachines[i].Shutdown(0, nil)
		}
	}
}

// func (hypervisor HypervisorStruct) Connect() (libvirt.VirConnection, error) {
// 	connection, err := libvirt.NewVirConnection("qemu+tcp://" + hypervisor.Host + "/system")
// 	return connection, err
// }
