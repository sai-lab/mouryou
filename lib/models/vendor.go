package models

//import libvirt "github.com/rgbkrk/libvirt-go"

type VendorStruct struct {
	Name            string                          `json:"name"`
	VirtualMachines map[string]VirtualMachineStruct `json:"virtual_machines"`
}

func (vendor *VendorStruct) Initialize() {
	for _, v := range vendor.VirtualMachines {
		v.Vendor = vendor

		// if i != 0 {
		// 	vendor.VirtualMachines[i].Shutdown(0, nil)
		// }
	}
}

// func (vendor vendorStruct) Connect() (libvirt.VirConnection, error) {
// 	connection, err := libvirt.NewVirConnection("qemu+tcp://" + vendor.Host + "/system")
// 	return connection, err
// }
