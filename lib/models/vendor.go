package models

//import libvirt "github.com/rgbkrk/libvirt-go"

type VendorStruct struct {
	Name            string                    `json:"name"`
	VirtualMachines map[string]VirtualMachine `json:"virtual_machines"`
}

// Initialize はVMに所属しているvendorの情報を追加します。
func (vendor *VendorStruct) Initialize() {
	for _, v := range vendor.VirtualMachines {
		v.Vendor = vendor

	}
}

// Connect はKVMホストへの接続を行います。
// 接続がエラーを出すようになったので一旦コメントアウトしています。
//func (vendor VendorStruct) Connect() (libvirt.VirConnection, error) {
//	connection, err := libvirt.NewVirConnection("qemu+tcp://" + vendor.Host + "/system")
//	return connection, err
//}
