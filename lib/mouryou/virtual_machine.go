package mouryou

import (
	"../apache"
	"time"
)

type virtualMachine struct {
	Name string
	Host string
	HV   *hypervisor
}

func (vm virtualMachine) operatingRatio() float64 {
	board, err := apache.Scoreboard(vm.Host)

	if err != nil {
		switch err.Error() {
		case "apache: no response":
			return 1.0
		case "apache: request timeout":
			return 1.0
		}
	}

	return apache.OperatingRatio(board)
}

func (vm virtualMachine) create(sleep time.Duration) {
	conn, err := vm.HV.connect()
	checkError(err)
	defer conn.CloseConnection()

	dom, err := conn.LookupDomainByName(vm.Name)
	checkError(err)

	dom.Create()
	time.Sleep(sleep * time.Second)

	writeOperating(false)
	powerCh <- "created"
}

func (vm virtualMachine) shutdown(sleep time.Duration) {
	conn, err := vm.HV.connect()
	checkError(err)
	defer conn.CloseConnection()

	dom, err := conn.LookupDomainByName(vm.Name)
	checkError(err)

	time.Sleep(sleep * time.Second)
	dom.Shutdown()

	writeOperating(false)
}
