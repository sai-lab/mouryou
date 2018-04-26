package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type VirtualMachine struct {
	Id         int               `json:"id"`
	Name       string            `json:"name"`
	Host       string            `json:"host"`
	Average    int               `json:"average"`
	Weight     int               `json:"weight"`
	Hypervisor *HypervisorStruct `json:"-"`
	Vendor     *VendorStruct     `json:"-"`
}

func (machine VirtualMachine) ServerState() apache.ServerStatus {
	var state apache.ServerStatus

	board, err := apache.Scoreboard(machine.Host)
	if err != nil {
		state.HostName = machine.Name
		state.Other = "Connection is timeout."
	} else {
		err = json.Unmarshal(board, &state)
		if err != nil {
			logger.PrintPlace(fmt.Sprint(err))
		}
	}
	state.Id = machine.Id

	return state
}

func (machine VirtualMachine) Bootup(sleep time.Duration) string {
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

	return "booted up"
}

func (machine VirtualMachine) Shutdown(sleep time.Duration) string {
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

	return "shutted down"
}
