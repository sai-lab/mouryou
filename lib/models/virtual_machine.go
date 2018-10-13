package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type VirtualMachine struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Host      string `json:"host"`
	Operation string `json:"operation"`
	// スループットの平均値
	Average int `json:"average"`
	// 現在の負荷状況(スループット基準) 0:普通 1:過負荷 2:低負荷
	LoadStatus int `json:"load_status"`
	// 基準の重さ
	BasicWeight int `json:"basic_weight"`
	// 現在の重さ
	Weight int `json:"weight"`
	// ハイパーバイザ
	Hypervisor *HypervisorStruct `json:"-"`
	// ベンダー
	Vendor *VendorStruct `json:"-"`
}

// ServerState はapache.Scoreboardから負荷状況を受け取り返却します。
func (machine VirtualMachine) ServerStatus() apache.ServerStatus {
	var status apache.ServerStatus

	board, err := apache.Scoreboard(machine.Host)
	if err != nil {
		// errがあった場合、timeoutしていると判断します。
		status.HostName = machine.Name
		status.Other = "Connection is timeout."
	} else {
		err = json.Unmarshal(board, &status)
		if err != nil {
			logger.PrintPlace(fmt.Sprint(err))
		}
	}
	status.Id = machine.Id

	return status
}

// Bootup はVMの起動処理を行います。
// 現在は実際に起動停止は行わないため起動にかかる時間分sleepします。
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

// Bootup はVMの起動処理を行います。
// 現在は実際に起動停止は行わないため停止にかかる時間分sleepします。
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

func ValidateOperation(str string) error {
	switch str {
	case "booting up":
	case "booted up":
	case "shutting down":
	case "shutted down":
	default:
		return errors.New("Operation Setting is invalid")

	}
	return nil
}
