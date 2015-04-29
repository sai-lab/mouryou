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

func (vm virtualMachine) create(sleep time.Duration, ch bool) {
	conn, err := vm.HV.connect()
	checkError(err)
	defer conn.CloseConnection()

	dom, err := conn.LookupDomainByName(vm.Name)
	checkError(err)

	dom.Create()
	time.Sleep(sleep * time.Second)
	if ch {
		powerCh <- "created"
	}
}

func (vm virtualMachine) shutdown(sleep time.Duration, ch bool) {
	time.Sleep(sleep * time.Second)

	conn, err := vm.HV.connect()
	checkError(err)
	defer conn.CloseConnection()

	dom, err := conn.LookupDomainByName(vm.Name)
	checkError(err)

	dom.Shutdown()
	if ch {
		powerCh <- "shutdowned"
	}
}
