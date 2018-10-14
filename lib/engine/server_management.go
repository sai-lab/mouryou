package engine

import (
	"sync"

	"fmt"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

// ServerManagement は起動状況と負荷状況に基いてオートスケールを実行します.
// 起動状況はengine.destination_settingが設定しています.
// 負荷状況はmonitor.LoadChから取得します.
func ServerManagement(config *models.Config) {
	var order Scale
	for order = range scaleCh {
		if config.DevelopLogLevel >= 3 {
			place := logger.Place()
			logger.Debug(place, fmt.Sprintf("Load: %s, Handle: %s, Weight: %d", order.Load, order.Handle, order.Weight))
		}
		switch order.Handle {
		case "ScaleOut":
			bootUpVMs(config, order.Weight)
		case "ScaleIn":
			shutDownVMs(config, order.Weight)
		default:
			place := logger.Place()
			logger.Debug(place, "Unknown Handle is comming!")
		}
	}
}

// bootUpVMs
func bootUpVMs(config *models.Config, weight int) {
	var candidate []int

	statuses := monitor.GetStates()

	for i, status := range statuses {
		// 停止中のサーバ以外は無視
		if status.Info != "shutted down" {
			continue
		}
		if status.Weight >= weight {
			go bootUpVM(config, status)
			mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+status.Weight)
			return
		}
		candidate = append(candidate, i)
	}

	if len(candidate) == 0 {
		return
	}

	boot := candidate[0]
	for _, n := range candidate {
		if statuses[n].Weight > statuses[boot].Weight {
			boot = n
		}
	}
	go bootUpVM(config, statuses[boot])
	mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+statuses[boot].Weight)
}

// bootUpVM
func bootUpVM(config *models.Config, st monitor.State) {
	var p monitor.PowerStruct

	p.Name = st.Name
	p.Info = "booting up"
	st.Info = "booting up"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, st.Name+" is booting up")
	}

	p.Info = config.Cluster.VirtualMachines[st.Name].Bootup(config.Sleep)
	st.Info = p.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, st.Name+" is boot up")
	}
	mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+st.Weight)
}

// shutDownVMs
func shutDownVMs(config *models.Config, weight int) {
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for _, st := range monitor.States {
		// 稼働中のサーバ以外は無視
		if st.Info != "booted up" {
			continue
		}
		// オリジンサーバは無視
		if config.ContainMachineName(config.OriginMachineNames, st.Name) {
			continue
		}
		// 常に稼働するサーバは無視
		if config.ContainMachineName(config.AlwaysRunningMachines, st.Name) {
			continue
		}

		if st.Weight <= weight {
			go shutDownVM(config, st)
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-st.Weight)
			mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight-st.Weight)
			if config.DevelopLogLevel >= 1 {
				place := logger.Place()
				logger.Debug(place, st.Name+" going to shutdown")
			}
			return
		}
	}
}

//shutDownVM
func shutDownVM(config *models.Config, st monitor.State) {
	var p monitor.PowerStruct
	p.Name = st.Name
	p.Info = "shutting down"
	st.Info = "shutting down"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}

	p.Info = config.Cluster.VirtualMachines[st.Name].Shutdown(config.Sleep)
	st.Info = p.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
}
