package mouryou

import (
	"github.com/alexzorin/libvirt-go"
)

type hypervisor struct {
	Host string
	VMs  []virtualMachine
}

func (hv *hypervisor) init() {
	for i := range hv.VMs {
		hv.VMs[i].HV = hv
		if i != 0 {
			hv.VMs[i].shutdown(0, false)
		}
	}
}

func (hv hypervisor) connect() (libvirt.VirConnection, error) {
	conn, err := libvirt.NewVirConnection("qemu+tcp://" + hv.Host + "/system")
	return conn, err
}
