package engine

import (
	"fmt"
	"sync"

	"time"

	"strconv"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func Initialize(config *models.Config) {
	for name, machine := range config.Cluster.VirtualMachines {
		var st monitor.State
		st.Name = name
		if config.DevelopLogLevel >= 4 {
			logger.PrintPlace("Machine ID: " + strconv.Itoa(machine.Id) + ", Machine Name: " + name)
		}

		err := config.Cluster.LoadBalancer.ChangeWeight(name, machine.Weight)
		if err != nil {
			fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
			break
		}
		st.Weight = machine.Weight
		if config.DevelopLogLevel >= 4 {
			logger.PrintPlace("Machine ID: " + strconv.Itoa(machine.Id) + ", Machine Name: " + name)
		}

		if config.ContainID(machine.Id) {
			if config.DevelopLogLevel > 1 {
				logger.PrintPlace("LogLevel 1 : set booted up " + " Machine Name: " + name +
					" Weight: " + strconv.Itoa(machine.Weight))
			}
			st.Info = "booted up"
			totalWeight += machine.Weight
			futureTotalWeight += machine.Weight
		} else {
			st.Info = "shutted down"
			if config.DevelopLogLevel > 1 {
				logger.PrintPlace("LogLevel 1 : set shutted down " + " Machine Name: " + name +
					" Weight: " + strconv.Itoa(machine.Weight))
			}
		}
		monitor.States = append(monitor.States, st)
	}
}

func WeightManager(config *models.Config) {
	for informations := range monitor.DataCh {
		for _, information := range informations {
			// エラーがあればdecreaseWeight, なければincreaseWeight
			// Connection is Timeout や Operating Ratio and CPU UsedPercent is MAX! など
			if !config.IsWeightChange {
				continue
			}
			if information.Error != "" {
				decreaseWeight(information, config)
			} else {
				increaseWeight(information, config)
			}
		}
	}
}

func decreaseWeight(information monitor.Data, config *models.Config) {
	var rwMutex sync.RWMutex
	name := information.Name

	rwMutex.RLock()
	for i, v := range monitor.States {
		if v.Name != name {
			continue
		}
		lowWeight := config.Cluster.VirtualMachines[v.Name].BasicWeight / 2
		basicWeight := config.Cluster.VirtualMachines[v.Name].BasicWeight
		// 重さがすでに下がっていれば break
		if v.Weight <= basicWeight-lowWeight {
			break
		}
		err := config.Cluster.LoadBalancer.ChangeWeight(information.Name, lowWeight)
		if err != nil {
			fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
		}

		// サーバの重みを変更したとき、合計の重みと最終的な重みを変更する
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-(monitor.States[i].Weight-lowWeight))
		mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight-(monitor.States[i].Weight-lowWeight))
		// 共有変数の重みを変更する
		monitor.States[i].Weight = lowWeight
		// 共有変数の変更時間を変更する
		monitor.States[i].Changed = time.Now()
		break
	}
	rwMutex.RUnlock()
}

func increaseWeight(information monitor.Data, config *models.Config) {
	var rwMutex sync.RWMutex
	name := information.Name

	rwMutex.RLock()
	for i, v := range monitor.States {
		// 名前が違う or 前回重さを変更した時間がconfig.RestorationTime秒より後なら continue
		if v.Name != name || v.Changed.After(time.Now().Add(time.Second*-config.RestorationTime)) {
			continue
		}
		basicWeight := config.Cluster.VirtualMachines[v.Name].BasicWeight
		if v.Weight >= basicWeight {
			break
		}

		err := config.Cluster.LoadBalancer.ChangeWeight(information.Name, basicWeight)
		if err != nil {
			fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
		}

		// サーバの重みを変更したとき、合計の重みと最終的な重みを変更する
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+(basicWeight-monitor.States[i].Weight))
		mutex.Write(&futureTotalWeight, &futureTotalWeightMutex, futureTotalWeight+(basicWeight-monitor.States[i].Weight))
		// 共有変数の重みを変更する
		monitor.States[i].Weight = basicWeight
		// 共有変数の変更時間を変更する
		monitor.States[i].Changed = time.Now()
		break
	}
	rwMutex.RUnlock()

}
