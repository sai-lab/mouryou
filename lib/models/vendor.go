package models

//import libvirt "github.com/rgbkrk/libvirt-go"

import (
	"github.com/sai-lab/mouryou/lib/logger"
)

type VendorStruct struct {
	Name            string                 `json:"name"`
	VirtualMachines []VirtualMachineStruct `json:"virtual_machines"`
}

func (vendor *VendorStruct) Initialize() {
	logger.PrintPlace("vendor Initialize")
	for i := range vendor.VirtualMachines {
		vendor.VirtualMachines[i].Vendor = vendor

		// if i != 0 {
		// 	vendor.VirtualMachines[i].Shutdown(0, nil)
		// }
	}
}

// func (vendor vendorStruct) Connect() (libvirt.VirConnection, error) {
// 	connection, err := libvirt.NewVirConnection("qemu+tcp://" + vendor.Host + "/system")
// 	return connection, err
// }
