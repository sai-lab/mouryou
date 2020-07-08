package models

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
)

type VirtualMachine struct {
	ID                   int               `json:"id"`
	Name                 string            `json:"name"`
	Host                 string            `json:"host"`
	Operation            string            `json:"operation"`
	ThroughputUpperLimit float64           `json:"throughput_upper_limit"` // スループットの平均値
	LoadStatus           int               `json:"load_status"`            // 現在の負荷状況(スループット基準) 0:普通 1:過負荷 2:低負荷
	BasicWeight          int               `json:"basic_weight"`           // 基準の重さ
	Weight               int               `json:"weight"`                 // 現在の重さ
	Hypervisor           *HypervisorStruct `json:"-"`                      // ハイパーバイザ
	Vendor               *VendorStruct     `json:"-"`                      // ベンダー
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
			place := logger.Place()
			logger.Error(place, err)
		}
	}
	status.Id = machine.ID

	return status
}

// SocketState はapache.Socketboardからソケット状況を受け取り返却します。
func (machine VirtualMachine) ServerStatus() apache.SocketStatus {
	var socket apache.SocketStatus

	board, err := apache.Socketboard(machine.Host)
	if err != nil {
		// errがあった場合、timeoutしていると判断します。
		socket.HostName = machine.Name
		socket.Other = "Connection is timeout."
	} else {
		err = json.Unmarshal(board, &socket)
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}
	}
	socket.Id = machine.ID

	return socket
}

// Bootup はVMの起動処理を行う．
// 現在は実際に起動停止は行わないため起動にかかる時間分sleepして擬似的な起動処理としている．
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
