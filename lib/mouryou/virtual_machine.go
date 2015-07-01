package mouryou

import (
	"time"

	"../apache"
)

type VirtualMachineStruct struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Hypervisor *HypervisorStruct
}

func (machine VirtualMachineStruct) OperatingRatio() float64 {
	board, err := apache.Scoreboard(machine.Host)
	if err != nil {
		return 1.0
	}

	return apache.OperatingRatio(board)
}

func (machine VirtualMachineStruct) Bootup(sleep int) {
	connection, err := machine.Hypervisor.Connect()
	if err != nil {
		powerCh <- err.Error()
		return
	}
	defer connection.CloseConnection()

	domain, err := connection.LookupDomainByName(machine.Name)
	if err != nil {
		powerCh <- err.Error()
		return
	}

	err = domain.Create()
	if err != nil {
		powerCh <- err.Error()
		return
	}

	time.Sleep(time.Duration(sleep) * time.Second)
	powerCh <- "bootup"
}

func (machine VirtualMachineStruct) Shutdown(sleep int) {
	powerCh <- "shutdown"

	connection, err := machine.Hypervisor.Connect()
	if err != nil {
		powerCh <- err.Error()
		return
	}
	defer connection.CloseConnection()

	domain, err := connection.LookupDomainByName(machine.Name)
	if err != nil {
		powerCh <- err.Error()
		return
	}

	time.Sleep(time.Duration(sleep) * time.Second)

	err = domain.Shutdown()
	if err != nil {
		powerCh <- err.Error()
		return
	}
}
