package engine

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/predictions"
)

func LoadDetermination(config *models.Config) {
	for {
		go ORBase(config)
		if config.UseThroughput {
			go TPBase(config)
		}
	}
}

func ORBase(config *models.Config) {
	var (
		// totalOR means the total value of the operating ratios of the working servers
		totalOR float64
		// working means the number of working servers
		working int
		// booting means the number of booting servers
		booting int
		// shutting means the number of servers that are stopped
		shutting int
		// lTotalWeight means the total value of the weights of the working servers
		lTotalWeight int
		// lFutureTotalWeight means the weight that will change due to bootup etc
		lFutureTotalWeight int
		// necessaryWeights means the necessary weights
		necessaryWeights int
		// orders is auto scale order
		orders []Scale
	)

	r := ring.New(LING_SIZE)
	ttlORs := make([]float64, LING_SIZE)

	for totalOR = range monitor.LoadORCh {
		r.Value = totalOR
		r = r.Next()
		ttlORs = convert.RingToArray(r)

		// Get Number of Active Servers
		working = mutex.Read(&working, &workMutex)
		booting = mutex.Read(&booting, &bootMutex)
		shutting = mutex.Read(&shutting, &shutMutex)
		lTotalWeight = mutex.Read(&totalWeight, &totalWeightMutex)
		lFutureTotalWeight = mutex.Read(&futureTotalWeight, &futureTotalWeightMutex)

		wServerManagementLog := []string{"engine.ORBase Log: ",
			fmt.Sprintf(
				"working: %d, booting: %d, shutting: %d, totalWeight: %d, futureTotalWeight: %d",
				working, booting, shutting, lTotalWeight, lFutureTotalWeight)}
		logger.Write(wServerManagementLog)

		// Exec Algorithm
		if config.UseHetero {
			necessaryWeights = predictions.ExecDifferentAlgorithm(config, working,
				booting, shutting, lTotalWeight, lFutureTotalWeight, ttlORs)
			switch {
			case necessaryWeights > lTotalWeight:
				orders = append(orders, Scale{Handle: "ScaleOut",
					Weight: necessaryWeights - lFutureTotalWeight, Load: "OR"})
			case necessaryWeights < lTotalWeight:
				orders = append(orders, Scale{Handle: "ScaleIn",
					Weight: necessaryWeights - lFutureTotalWeight, Load: "OR"})
			}
		} else {
			orders = scaleSameServers(config, ttlORs, working, booting,
				shutting, lTotalWeight, lFutureTotalWeight)
		}

		for _, order := range orders {
			scaleCh <- order
		}
	}
}

// scaleSameServersは単一性能向けアルゴリズムのサーバ起動停止メソッドです.
// predictions.ExecSameAlgorithmメソッドからmodels.Config.Sleep時間後に必要な台数と
// スケールインするかの真偽値を受け取り,それらに従って起動停止処理を実行します.
func scaleSameServers(c *models.Config, ttlORs []float64, working int, booting int,
	shutting int, tw int, fw int) []Scale {
	var (
		scaleIn        bool
		requiredNumber float64
		orders         []Scale
	)

	requiredNumber, scaleIn = predictions.ExecSameAlgorithm(c, working, booting, shutting, tw, fw, ttlORs)
	if c.DevelopLogLevel >= 1 {
		logger.PrintPlace("required server num is " + strconv.Itoa(int(requiredNumber)))
	}

	switch {
	case working+booting < len(c.Cluster.VirtualMachines) && int(requiredNumber) > 0 && shutting == 0:
		i := 0
		for i = 0; i < int(requiredNumber) && working+booting+i < len(c.Cluster.VirtualMachines); i++ {
			orders = append(orders, Scale{Handle: "ScaleOut", Weight: 10, Load: "OR"})
		}
	case working > 1 && scaleIn && mutex.Read(&waiting, &waitMutex) == 0 && booting == 0:
		orders = append(orders, Scale{Handle: "ScaleIn", Weight: 10, Load: "OR"})
	}

	return orders
}

func TPBase(config *models.Config) {
	needScaleOut := false
	needScaleIn := false

	for name, value := range config.Cluster.VirtualMachines {
		value.LoadStatus = judgeEachStatus(name, value.Average, config)
		// 0:普通 1:高負荷 2:低負荷
		switch value.LoadStatus {
		case 1:
			needScaleOut = true
		case 2:
			needScaleIn = true
		default:
		}
	}

	if needScaleOut {
		scaleCh <- Scale{Handle: "ScaleOut", Weight: 10, Load: "TP"}
	} else if needScaleIn {
		scaleCh <- Scale{Handle: "ScaleIn", Weight: 10, Load: "TP"}
	}
}

// スループットを用いて各サーバの負荷状況を判断する
// 0:普通 1:高負荷 2:低負荷
func judgeEachStatus(serverName string, average int, config *models.Config) int {
	var val float64
	var twts [30]models.ThroughputWithTime

	query := "SELECT time, throughput FROM " + config.InfluxDBServerDB + " WHERE host = '" + serverName + "' LIMIT 30"
	res, err := databases.QueryDB(config.InfluxDBConnection, query, config.InfluxDBPasswd)
	if err != nil {
		logger.WriteMonoString(err.Error())
	}

	if res == nil {
		return 0
	}
	for i, row := range res[0].Series[0].Values {
		t, err := time.Parse(time.RFC3339, row[0].(string))
		if err != nil {
			logger.WriteMonoString(err.Error())
		}
		val, err = row[1].(json.Number).Float64()
		if err != nil {
			logger.WriteMonoString(err.Error())
		}
		twts[i] = models.ThroughputWithTime{val, t}
	}

	if judgeHighLoadByThroughput(config, serverName, twts) {
		return 1
	}
	if judgeLowLoadByThroughput(config, serverName, twts) {
		return 2
	}
	return 0
}

// 規定回数以上上限スループットを超えれば高負荷と判断
func judgeHighLoadByThroughput(config *models.Config, serverName string, twts [30]models.ThroughputWithTime) bool {
	TPHigh := config.Cluster.VirtualMachines[serverName].Average
	c := 0
	for i, twt := range twts {
		fmt.Println(twt.Throughput, twt.MeasurementTime)
		if twt.Throughput > float64(TPHigh) {
			c++
		}
		if i+1 == config.ThroughputScaleOutTime {
			break
		}
	}

	if c >= config.ThroughputScaleOutThreshold {
		return true
	}
	return false
}

// 規定回数以上連続して下限スループットを下回れば低負荷と判断
func judgeLowLoadByThroughput(config *models.Config, serverName string, twts [30]models.ThroughputWithTime) bool {
	TPHigh := config.Cluster.VirtualMachines[serverName].Average
	rate := config.ThroughputScaleInRate
	c := 0
	for i, twt := range twts {
		fmt.Println(twt.Throughput, twt.MeasurementTime)
		if twt.Throughput > float64(TPHigh)*rate {
			c++
		} else {
			c = 0
		}
		if i+1 == config.ThroughputScaleInTime {
			break
		}
	}

	if c >= config.ThroughputScaleInThreshold {
		return true
	}
	return false
}
