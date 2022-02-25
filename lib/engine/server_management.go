package engine

import (
	"fmt"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

// ServerManagement は稼働状況と負荷状況に基いてオートスケールを実行します.
// 稼働状況はengine/destination_setting.goが設定しています.
// 負荷状況はengine/load_determination.goが判断します.
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
		if config.Cluster.VirtualMachines[serverState.Name].Role == "pool" {
			// poolサーバは無視
			continue
		}
		if serverState.Info != "shutted down" {
			// 停止中のサーバ以外は無視
			continue
		}

		if serverState.Weight >= weight {
			// サーバの重さが必要な重み以上なら起動処理を任せてreturn
			go bootUpVMpool(config, serverState, load)
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
	go bootUpVMpool(config, serverStates[toBootUp], load)
	mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+serverStates[toBootUp].Weight)
}

func bootUpVMpool(config *models.Config, serverState monitor.ServerState, load string) {
	var power monitor.PowerStruct
	var pp monitor.PowerStruct
	var pool monitor.ServerState

	// これから起動処理を発行することを通知
	power.Name = serverState.Name
	power.Info = "booting up"
	power.Load = load
	serverState.Info = "booting up"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- serverState
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, serverState.Name+" is booting up")
	}

	pool.Name = ""
	/*
		for _, status := range monitor.GetStates() {
			if status.Info == "shutted down" && status.Name != st.Name {
				pool.Name = status.Name
			}
		}
	*/
	// プールサーバ
	for _, status := range monitor.GetServerStates() {
		if status.Info == "shutted down" && config.Cluster.VirtualMachines[status.Name].Role == "pool" {
			pool.Name = status.Name
		}
	}
	if pool.Name != "" {
		pool.Weight = 5
		pp.Name = pool.Name
		pp.Info = "booting up"
		pp.Load = load
		pool.Info = "booting up"
		if monitor.PowerCh != nil {
			monitor.PowerCh <- pp
		}
		if monitor.StateCh != nil {
			monitor.StateCh <- pool
		}
		if config.DevelopLogLevel >= 1 {
			//logger.PrintPlace("BootUp " + pool.Name)
			//fmt.Println("Pool   " + pool.Name + " is booting up")
			place := logger.Place()
			logger.Debug(place, pool.Name+" is booting up")
		}
		pp.Info = config.Cluster.VirtualMachines[pool.Name].Bootup(config.Start)
		pool.Info = "booted up"
		if monitor.PowerCh != nil {
			monitor.PowerCh <- pp
		}
		if monitor.StateCh != nil {
			monitor.StateCh <- pool
		}
		if config.DevelopLogLevel >= 1 {
			//logger.PrintPlace("BootUp " + pool.Name)
			//fmt.Println("Pool   " + pool.Name + " is booted up")
			place := logger.Place()
			logger.Debug(place, serverState.Name+" is booted up")
		}
		power.Info = config.Cluster.VirtualMachines[serverState.Name].Bootup(config.Sleep - config.Start)
	} else {
		// 起動処理を発行，完了後の返却値受け取り
		power.Info = config.Cluster.VirtualMachines[serverState.Name].Bootup(config.Sleep)
	}
	serverState.Info = power.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- serverState
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, serverState.Name+" is boot up")
	}

	if pool.Name != "" {
		//i := 10
		pp.Info = config.Cluster.VirtualMachines[pool.Name].Shutdown(config.Sleep + config.Stop)
		/*
			fmt.Println("## PoolChangeWeight")
			for i > 1 {
				fmt.Println("### PoolChangeWeight")
				PoolFireChangeWeight(c, pool.Name, i)
				fmt.Println(i)
				time.Sleep(3 * time.Second)
				i--
			}
		*/
		pp.Info = "shutting down"
		pool.Info = pp.Info
		if monitor.PowerCh != nil {
			monitor.PowerCh <- pp
		}
		if monitor.StateCh != nil {
			monitor.StateCh <- pool
		}
		if config.DevelopLogLevel >= 1 {
			//fmt.Println("Pool   " + pool.Name + " is shutting down")
			place := logger.Place()
			logger.Debug(place, pool.Name+" is shutting up")
		}

		pp.Info = "shutted down"
		pool.Info = pp.Info
		if monitor.PowerCh != nil {
			monitor.PowerCh <- pp
		}
		if monitor.StateCh != nil {
			monitor.StateCh <- pool
		}
		if config.DevelopLogLevel >= 1 {
			//fmt.Println("Pool   " + pool.Name + " is shutted down")
			place := logger.Place()
			logger.Debug(place, pool.Name+" is shutted up")
		}

		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-pool.Weight)
		if config.DevelopLogLevel >= 1 {
			fmt.Println(pool.Name + " going to shutdown")
		}
	} else {
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+serverState.Weight)
	}
}

// bootUpVM は引数に 設定値用構造体 config, 起動するサーバの情報 serverState, 判断基準にした負荷量 load をとります．
func bootUpVM(config *models.Config, serverState monitor.ServerState, load string) {
	var power monitor.PowerStruct

	// これから起動処理を発行することを通知
	power.Name = serverState.Name
	power.Info = "booting up"
	power.Load = load
	serverState.Info = "booting up"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- serverState
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, serverState.Name+" is booting up")
	}

	// 起動処理を発行，完了後の返却値受け取り
	power.Info = config.Cluster.VirtualMachines[serverState.Name].Bootup(config.Sleep)
	serverState.Info = power.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- serverState
	}
	if config.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, serverState.Name+" is boot up")
	}
	mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+serverState.Weight)
}

// shutDownVMs は引数に 設定値用構造体 config, 停止したいサーバの重み weight, 判断基準にした負荷量 load をとります．
func shutDownVMs(config *models.Config, weight int, load string) {
	for _, serverState := range monitor.GetServerStates() {
		// 稼働中のサーバ以外は無視
		if serverState.Info != "booted up" {
			continue
		}
		// オリジンサーバは無視
		if config.ContainMachineName(config.OriginMachineNames, serverState.Name) {
			continue
		}
		// 常に稼働するサーバは無視
		if config.ContainMachineName(config.AlwaysRunningMachines, serverState.Name) {
			continue
		}

		if serverState.Weight <= weight {
			// サーバの重さが必要な重み以下なら停止処理を発行
			go shutDownVM(config, serverState, load)
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-serverState.Weight)
			mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight-serverState.Weight)
			if config.DevelopLogLevel >= 1 {
				place := logger.Place()
				logger.Debug(place, serverState.Name+" going to shutdown")
			}
			return
		}
	}
}

//shutDownVM は引数に 設定値用構造体 config, 停止するサーバの情報 serverState, 判断基準にした負荷量 load をとります．
func shutDownVM(config *models.Config, serverState monitor.ServerState, load string) {
	var power monitor.PowerStruct
	power.Name = serverState.Name
	power.Info = "shutting down"
	power.Load = load
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}

	serverState.Info = "shutting down"
	if monitor.StateCh != nil {
		monitor.StateCh <- serverState
	}

	// 停止処理を発行，完了後の返却値受け取り
	power.Info = config.Cluster.VirtualMachines[serverState.Name].Shutdown(config.Sleep)
	if monitor.PowerCh != nil {
		monitor.PowerCh <- power
	}

	serverState.Info = power.Info
	if monitor.StateCh != nil {
		monitor.StateCh <- serverState
	}
}
