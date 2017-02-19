package models

import (
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
	UnitTime   time.Duration     `json:"unit_time"`
	UnitCost   float64           `json:"unit_cost"`
	CostCount  int
	BootTime   time.Time
}

func (machine VirtualMachineStruct) OperatingRatio() float64 {
	board, err := apache.Scoreboard(machine.Host)
	if err != nil {
		return 1.0
	}

	return apache.OperatingRatio(board)
}

func (machine VirtualMachineStruct) Bootup(sleep time.Duration, power chan string) {
	if power != nil {
		power <- "booting up"
	}
	machine.BootTime = time.Now()
	machine.CostCount = 0
	select {
	case vmQue[machine.Id] <- false:
	default:
	}

	go func() {
		t := time.NewTicker((machine.UnitTime + 1) * time.Second)
		for {
			select {
			case <-t.C:
				machine.CostCount++
				logger.Write("CostInfo, " + fmt.Sprint(machine.UnitCost))
				logger.Print("CostInfo, " + fmt.Sprint(machine.UnitCost))
			case stop := <-vmQue[machine.Id]:
				if stop {
					logger.PrintPlace("stop ticker")
					t.Stop()
					vmQue[machine.Id] <- false
				}
			default:
			}
		}
	}()

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
}

func (machine VirtualMachineStruct) Shutdown(sleep time.Duration, power chan string) {
	if power != nil {
		power <- "shutting down"
	}

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
	select {
	case vmQue[machine.Id] <- true:
	default:
	}
}
