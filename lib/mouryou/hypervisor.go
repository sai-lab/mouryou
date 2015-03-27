package mouryou

type hypervisor struct {
	Host string
	VMs  []virtualMachine
}

func (hv *hypervisor) assignVMs() {
	for _, vm := range hv.VMs {
		vm.HV = hv
	}
}
