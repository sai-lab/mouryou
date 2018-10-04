package engine

import (
	"fmt"
	"github.com/sai-lab/mouryou/lib/db"
	"github.com/sai-lab/mouryou/lib/models"
)

func ThroughputBaseAlgorithm(config models.Config) {

	for name, value := range config.Cluster.VirtualMachines {
		value.LoadStatus = judgeEachStatus(name, value.Average, config)
	}
}

// スループットを用いて各サーバの負荷状況を判断する
// 0:普通 1:過負荷 2:低負荷
func judgeEachStatus(serverName string, average int, config models.Config) int {
	throughputsWithTimes := db.LoadThroughput(serverName)
	if judgeHighLoadByThroughput(config, serverName, throughputsWithTimes) {
		return 1
	}
	if judgeLowLoadByThroughput(config, serverName, throughputsWithTimes) {
		return 2
	}
	return 0
}

// 規定回数以上上限スループットを超えれば過負荷と判断
func judgeHighLoadByThroughput(config models.Config, serverName string, twts []models.ThroughputWithTime) bool {
	TPHigh := config.Cluster.VirtualMachines[serverName].Average
	c := 0
	for throughput, time := range twts {
		fmt.Println(throughput, time)
		if throughput > TPHigh {
			c++
		}
	}

	if c >= config.ThroughputScaleOutThreshold {
		return true
	}
	return false
}

// 規定回数以上連続して上限スループットを超えれば過負荷
func judgeLowLoadByThroughput(config models.Config, serverName string, twts []models.ThroughputWithTime) bool {
	TPHigh := config.Cluster.VirtualMachines[serverName].Average
	rate := config.ThroughputScaleInRate
	c := 0
	for throughput, time := range twts {
		fmt.Println(throughput, time)
		if throughput > int(float64(TPHigh)*rate) {
			c++
		} else {
			c = 0
		}
	}

	if c >= config.ThroughputScaleInThreshold {
		return true
	}
	return false
}
