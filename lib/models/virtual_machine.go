package models

import (
	"fmt"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type VirtualMachineStruct struct {
	Name       string            `json:"name"`
	Host       string            `json:"host"`
	Hypervisor *HypervisorStruct `json:"-"`
	Vendor     *VendorStruct     `json:"-"`
}

func (machine VirtualMachineStruct) ServerState() *ServerStat {
	var state ServerStat

	board, err := apache.Scoreboard(machine.Host)
	if err != nil {
		logger.PrintPlace(fmt.Sprint(err))
	}

	err := json.Unmarshal(board, &state)
	if err != nil {
		return 0
	}

	return state
}

func (machine VirtualMachineStruct) Bootup(sleep time.Duration, power chan string) {
	if power != nil {
		power <- "booting up"
	}
	logger.PrintPlace("Booting up")

	// connection, err := machine.Hypervisor.Connect()
	// if err != nil {
	// 	power <- err.Error()
	// 	return
	// }
	// defer connection.CloseConnection()

	// domain, err := connection.LookupDomainByName(machine.Name)
	// if err != nil {
	// 	power <- err.Error()
	// 	return
	// }

	// err = domain.Create()
	// if err != nil {
	// 	power <- err.Error()
	// 	return
	// }

	time.Sleep(sleep * time.Second)

	if power != nil {
		power <- "booted up"
	}
	logger.PrintPlace("Booted up")
}

func (machine VirtualMachineStruct) Shutdown(sleep time.Duration, power chan string) {
	if power != nil {
		power <- "shutting down"
	}
	logger.PrintPlace("Virtual Machine Shutdown")

	// connection, err :=  machine.Hypervisor.Connect() // here?

	// if err != nil {
	// 	power <- err.Error()
	// 	return
	// }
	// defer connection.CloseConnection()

	// domain, err := connection.LookupDomainByName(machine.Name)
	// if err != nil {
	// 	power <- err.Error()
	// 	return
	// }

	// time.Sleep(sleep * time.Second)
	// err = domain.Shutdown()
	// if err != nil {
	// 	power <- err.Error()
	// 	logger.PrintPlace(fmt.Sprint(err.Error))
	// 	return
	// }

	time.Sleep(sleep * time.Second)

	if power != nil {
		power <- "shutted down"
	}
}
