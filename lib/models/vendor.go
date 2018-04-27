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

		// オリジンサーバ(i==0)以外は停止する処理です。
		// 現在は利用していないためコメントアウトしています。
		// if i != 0 {
		// 	vendor.VirtualMachines[i].Shutdown(0, nil)
		// }
	}
}

// Connect はKVMホストへの接続を行います。
// 接続がエラーを出すようになったので一旦コメントアウトしています。
//func (vendor VendorStruct) Connect() (libvirt.VirConnection, error) {
//	connection, err := libvirt.NewVirConnection("qemu+tcp://" + vendor.Host + "/system")
//	return connection, err
//}
