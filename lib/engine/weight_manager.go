package engine

import (
	"fmt"
	"time"

	"strconv"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

// ServerWeightInitialize はmouryou起動時にmouryou.jsonから得た情報を元に各サーバの重さを初期化します．
func ServerWeightInitialize(config *models.Config, startNum int) {
	// 起動中と認識している台数をstartNumの値で更新
	mutex.Write(&working, &workMutex, startNum)
	for name, machine := range config.Cluster.VirtualMachines {
		var st monitor.ServerState // サーバの状態を格納する構造体
		st.Name = name
		if config.DevelopLogLevel > 5 {
			place := logger.Place()
			logger.Debug(place, "Machine ID: "+strconv.Itoa(machine.ID)+", Machine Name: "+name)
		}

		// mouryou.jsonで指定した重さでロードバランサに登録
		err := config.Cluster.LoadBalancer.ChangeWeight(name, machine.Weight)
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
			break
		}

		// ログとデータベースに記録
		tags := []string{
			"operation:change_weight",
			fmt.Sprintf("host:%s", machine.Name),
		}
		fields := []string{
			fmt.Sprintf("weight:%d", machine.Weight),
		}
		logger.Record(tags, fields)
		databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

		// 登録完了後，stateの重さを更新
		st.Weight = machine.Weight
		if config.DevelopLogLevel > 5 {
			place := logger.Place()
			logger.Debug(place, "Machine ID: "+strconv.Itoa(machine.ID)+", Machine Name: "+name+", Machine Weight:"+fmt.Sprint(st.Weight))
		}

		if config.IsStartMachineID(machine.ID) {
			if config.DevelopLogLevel > 1 {
				place := logger.Place()
				logger.Debug(place, "set booted up Machine Name: "+name+" Weight: "+strconv.Itoa(machine.Weight))
			}
			st.Info = "booted up"
			// 初期化処理中で他のメソッドはtotalWeightとfutureTotalWeightを参照していないため，
			// ここではMutex処理を行わずに代入
			totalWeight += machine.Weight
			futureTotalWeight += machine.Weight
		} else {
			st.Info = "shutted down"
			if config.DevelopLogLevel > 1 {
				place := logger.Place()
				logger.Debug(place, "set shutted down Machine Name: "+name+" Weight: "+strconv.Itoa(machine.Weight))
			}
		}
		// monitorで管理しているServerStatesに追加
		err = monitor.AddServerState(st)
		if err != nil {
			logger.Error(logger.Place(), err)
		}
	}
}

// 重さの操作を取り扱うメソッド
// タイムアウトなどのエラーがあれば重さを減らして負荷を低下させ，通常なら元の重さに戻す
func WeightManager(config *models.Config) {
	for conditions := range monitor.ConditionCh {
		for _, condition := range conditions {
			// エラーがあればdecreaseWeight, なければincreaseWeight
			// Connection is Timeout や Operating Ratio and CPU UsedPercent is MAX! など
			if !config.IsWeightChange {
				continue
			}
			if condition.Error != "" {
				decreaseWeight(condition, config)
			} else {
				increaseWeight(condition, config)
			}
		}
	}
}

func decreaseWeight(information monitor.Condition, config *models.Config) {
	for _, serverState := range monitor.GetServerStates() {
		if serverState.Name != information.Name {
			continue
		}
		lowWeight := config.Cluster.VirtualMachines[serverState.Name].BasicWeight / 2
		basicWeight := config.Cluster.VirtualMachines[serverState.Name].BasicWeight
		// 重さがすでに下がっていれば break
		if serverState.Weight <= basicWeight-lowWeight {
			break
		}
		err := config.Cluster.LoadBalancer.ChangeWeight(information.Name, lowWeight)
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}

		// サーバの重みを変更したとき、合計の重みと最終的な重みを変更する
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-(serverState.Weight-lowWeight))
		mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight-(serverState.Weight-lowWeight))
		// 共有変数の重みと変更時間を更新
		monitor.UpdateServerStates(information.Name, lowWeight, "", time.Now(), serverState.WaitTime)
		break
	}
}

func increaseWeight(information monitor.Condition, config *models.Config) {
	for _, serverState := range monitor.GetServerStates() {
		// 名前が違う or 前回重さを変更した時間がconfig.RestorationTime秒より後なら continue
		if serverState.Name != information.Name || serverState.Changed.After(time.Now().Add(time.Second*-config.RestorationTime)) {
			continue
		}
		basicWeight := config.Cluster.VirtualMachines[serverState.Name].BasicWeight
		if serverState.Weight >= basicWeight {
			break
		}

		err := config.Cluster.LoadBalancer.ChangeWeight(information.Name, basicWeight)
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}

		// サーバの重みを変更したとき、合計の重みと最終的な重みを変更する
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+(basicWeight-serverState.Weight))
		mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+(basicWeight-serverState.Weight))
		// 共有変数の重みと更新時間を変更する
		monitor.UpdateServerStates(information.Name, basicWeight, "", time.Now(), serverState.WaitTime)
		break
	}
}
