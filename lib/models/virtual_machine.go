package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type VirtualMachineStruct struct {
	Id         int               `json:"id"`
	Name       string            `json:"name"`
	Host       string            `json:"host"`
	Hypervisor *HypervisorStruct `json:"-"`
	Vendor     *VendorStruct     `json:"-"`
}

func (machine VirtualMachineStruct) ServerState() apache.ServerStat {
	var state apache.ServerStat

	board, err := apache.Scoreboard(machine.Host)
	if err != nil {
		//logger.PrintPlace("Scoreboard error! : " + fmt.Sprint(err))
		state.HostName = machine.Name
		state.Other = "Connection is timeout."
	} else {
		err = json.Unmarshal(board, &state)
		if err != nil {
			logger.PrintPlace(fmt.Sprint(err))
		}
	}

	return state
}

func (machine VirtualMachineStruct) Bootup(sleep time.Duration) string {
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

func (machine VirtualMachineStruct) Shutdown(sleep time.Duration) string {
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

// func (machine *VirtualMachineStruct) SetStatus(s string) {
// 	logger.PrintPlace("name: " + machine.Name + ", status: " + machine.Status + ", s: " + s)
// 	machine.Status = s
// 	logger.PrintPlace("name: " + machine.Name + ", status: " + machine.Status + ", s: " + s)
// }

// func (machine VirtualMachineStruct) SetWeight(w float64) {
// 	logger.PrintPlace("name: " + machine.Name + ", weight: " + fmt.Sprint(machine.Weight) + ", w: " + fmt.Sprint(w))
// 	machine.Weight = w
// 	logger.PrintPlace("name: " + machine.Name + ", weight: " + fmt.Sprint(machine.Weight) + ", w: " + fmt.Sprint(w))
// }
