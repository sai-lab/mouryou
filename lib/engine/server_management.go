package engine

import (
	"fmt"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

// ServerManagement は起動状況と負荷状況に基いてオートスケールを実行します.
// 起動状況はengine.destination_settingが設定しています.
// 負荷状況はmonitor.LoadChから取得します.
func ServerManagement(config *models.Config) {
	var b, s, w, wait int
	var order autoScaleOrder

	vmNum := len(config.Cluster.VirtualMachines)
	arm := len(config.AlwaysRunningMachines)

	for order = range autoScaleOrderCh {
		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shutting, &shutMutex)
		wait = mutex.Read(&waiting, &waitMutex)

		tags := []string{"parameter:working_log", "operation:server_management"}
		fields := []string{fmt.Sprintf("working:%d", w),
			fmt.Sprintf("booting:%d", b),
			fmt.Sprintf("shutting:%d", s),
			fmt.Sprintf("load:%s", order.Load),
			fmt.Sprintf("handle:%s", order.Handle),
			fmt.Sprintf("weight:%d", order.Weight),
		}
		logger.Record(tags, fields)
		databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

		if config.DevelopLogLevel >= 3 {
			place := logger.Place()
			logger.Debug(place, fmt.Sprintf("Load: %s, Handle: %s, Weight: %d", order.Load, order.Handle, order.Weight))
		}
		switch order.Handle {
		case "ScaleOut":
			if w+b < vmNum && s == 0 {
				bootUpVMs(config, order.Weight, order.Load)
			}
		case "ScaleIn":
			if w > arm && b == 0 && wait == 0 {
				shutDownVMs(config, order.Weight, order.Load)
			}
		default:
			place := logger.Place()
			logger.Debug(place, "Unknown Handle is comming!")
		}
	}
}

// bootUpVMs は引数に 設定値用構造体 config, 起動したいサーバの重み weight, 判断基準にした負荷量 load をとります．
// 複数形だけど一台ずつ起動処理をbootUpVMに投げます．
func bootUpVMs(config *models.Config, weight int, load string) {
	// 充分な重みを持つ単体のサーバが存在しない場合に，起動する候補となる重みの小さいサーバの添字を格納する配列
	var candidate []int
	serverStates := monitor.GetServerStates()

	for i, serverState := range serverStates {
		if serverState.Info != "shutted down" {
			// 停止中のサーバ以外は無視
			continue
		}

		if serverState.Weight >= weight {
			// サーバの重さが必要な重み以上なら起動処理を任せてreturn
			go bootUpVM(config, serverState, load)
			mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+serverState.Weight)
			return
		}
		// サーバの重さが必要な重み未満の場合candidateに追加
		candidate = append(candidate, i)
	}

	if len(candidate) == 0 {
		// 起動候補が存在しない場合何もせずreturn
		return
	}

	// 起動候補サーバの中から最も重みの大きいサーバを起動
	toBootUp := candidate[0]
	for _, n := range candidate {
		if serverStates[n].Weight > serverStates[toBootUp].Weight {
			toBootUp = n
		}
	}
	go bootUpVM(config, serverStates[toBootUp], load)
	mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+serverStates[toBootUp].Weight)
}

// bootUpVM
func bootUpVM(config *models.Config, st monitor.ServerState, load string) {
	var power monitor.PowerStruct

	power.Name = st.Name
	power.Info = "booting up"
	power.Load = load
	st.Info = "booting up"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, st.Name+" is booting up")
	}

	power.Info = config.Cluster.VirtualMachines[st.Name].Bootup(config.Sleep)
	st.Info = power.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
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
func shutDownVMs(config *models.Config, weight int, load string) {
	for _, st := range monitor.GetServerStates() {
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
			go shutDownVM(config, st, load)
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
func shutDownVM(config *models.Config, st monitor.ServerState, load string) {
	var p monitor.PowerStruct
	p.Name = st.Name
	p.Info = "shutting down"
	p.Load = load
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
