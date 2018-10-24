package engine

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"time"

	"strconv"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/predictions"
)

func LoadDetermination(config *models.Config) {
	if config.UseOperatingRatio {
		go operatingRatioBase(config)
	}
	if config.UseThroughput {
		go throughputBase(config)
	}
}

func operatingRatioBase(config *models.Config) {
	var (
		// totalOR means the total value of the operating ratios of the working servers
		totalOR float64
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
		lWorking := mutex.Read(&working, &workMutex)
		lBooting := mutex.Read(&booting, &bootMutex)
		lShutting := mutex.Read(&shutting, &shutMutex)
		lTotalWeight := mutex.Read(&totalWeight, &totalWeightMutex)
		lFutureTotalWeight := mutex.Read(&futureTotalWeight, &futureTotalWeightMutex)

		tags := []string{"base_load:or", "parameter:working_log"}
		fields := []string{fmt.Sprintf("working:%d", lWorking),
			fmt.Sprintf("booting:%d", lBooting),
			fmt.Sprintf("shutting:%d", lShutting),
			fmt.Sprintf("total_weight:%d", lTotalWeight),
			fmt.Sprintf("future_total_weight:%d", lFutureTotalWeight),
		}
		logger.Record(tags, fields)
		databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

		// Exec Algorithm
		if config.UseHetero {
			necessaryWeights = predictions.ExecDifferentAlgorithm(config, lWorking,
				lBooting, lShutting, lTotalWeight, lFutureTotalWeight, ttlORs)
			switch {
			case necessaryWeights > lTotalWeight:
				orders = append(orders, Scale{Handle: "ScaleOut",
					Weight: necessaryWeights - lFutureTotalWeight, Load: "OR"})
			case necessaryWeights < lTotalWeight:
				orders = append(orders, Scale{Handle: "ScaleIn",
					Weight: necessaryWeights - lFutureTotalWeight, Load: "OR"})
			}
		} else {
			orders = scaleSameServers(config, ttlORs, lWorking, lBooting,
				lShutting, lTotalWeight, lFutureTotalWeight)
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
	if c.DevelopLogLevel >= 2 {
		place := logger.Place()
		logger.Debug(place, fmt.Sprint("required server num is ", requiredNumber))
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

func throughputBase(config *models.Config) {
	for _ = range monitor.LoadTPCh {
		needScaleOut := false
		needScaleIn := false

		var bootedServers []string
		var bootingServers []string
		//
		for _, v := range monitor.GetStates() {
			if config.DevelopLogLevel >= 6 {
				place := logger.Place()
				logger.Debug(place, "GetStates() Machine Name: "+v.Name+"Machine Info: "+v.Info)
			}
			if v.Info == "booted up" {
				bootedServers = append(bootedServers, v.Name)
			} else if v.Info == "booting up" {
				bootingServers = append(bootingServers, v.Name)
			}
		}

		switch config.ThroughputAlgorithm {
		case "All":
			for _, name := range bootedServers {
				value := config.Cluster.VirtualMachines[name]

				value.LoadStatus = judgeEachStatus(name, value.ThroughputUpperLimit, config)
				// 0:普通 1:高負荷 2:低負荷
				switch value.LoadStatus {
				case 1:
					needScaleOut = true
				case 2:
					needScaleIn = true
				default:
					needScaleIn = false
					needScaleOut = false
				}
			}
		case "MovingAverage":
			totalTPRatioMovingAverage := 0.0
			for _, name := range bootedServers {
				value := config.Cluster.VirtualMachines[name]
				totalTPRatioMovingAverage += movingAverageOfThroughputRatio(name, value.ThroughputUpperLimit, config)
				tags := []string{"parameter:working_log", "operation:load_determination"}
				fields := []string{
					fmt.Sprintf("throughput_upper_limit:%f", value.ThroughputUpperLimit),
					fmt.Sprintf("total_of_moving_average_of_throughput_ratio:%f", totalTPRatioMovingAverage),
				}
				logger.Record(tags, fields)
				databases.WriteValues(config.InfluxDBConnection, config, tags, fields)
			}
			ratioAverage := totalTPRatioMovingAverage / float64(len(bootedServers))

			// ログにratioAverageを記録
			tags := []string{"parameter:working_log", "operation:load_determination"}
			fields := []string{
				fmt.Sprintf("ratio_average:%f", ratioAverage),
			}
			logger.Record(tags, fields)
			databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

			// 判定
			if ratioAverage >= config.ThroughputScaleOutRatio {
				needScaleOut = true
			} else if ratioAverage <= config.ThroughputScaleInRatio {
				needScaleIn = true
			}
		case "MovingAverageV1.2":
			totalThroughput := 0.0
			totalUpperLimit := 0.0
			for _, name := range bootedServers {
				value := config.Cluster.VirtualMachines[name]
				totalThroughput += intervalThroughputTotal(name, config)
				totalUpperLimit += value.ThroughputUpperLimit
			}
			for _, name := range bootingServers {
				// 起動処理中のサーバの上限スループットも考慮
				totalUpperLimit += config.Cluster.VirtualMachines[name].ThroughputUpperLimit
			}
			movingAverage := totalThroughput / (float64(config.ThroughputMovingAverageInterval) * totalUpperLimit)

			// ログにmovingAverageを記録
			tags := []string{"parameter:working_log", "operation:load_determination"}
			fields := []string{
				fmt.Sprintf("moving_average:%f", movingAverage),
			}
			logger.Record(tags, fields)
			databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

			// 判定
			if movingAverage >= config.ThroughputScaleOutRatio {
				needScaleOut = true
			} else if movingAverage <= config.ThroughputScaleInRatio {
				needScaleIn = true
			}
		default:
			panic("unknown algorithm")
		}
		if needScaleOut {
			scaleCh <- Scale{Handle: "ScaleOut", Weight: 10, Load: "TP"}
		} else if needScaleIn {
			scaleCh <- Scale{Handle: "ScaleIn", Weight: 10, Load: "TP"}
		}
	}
}

func intervalThroughputTotal(serverName string, config *models.Config) float64 {
	query := "SELECT time, throughput FROM " + config.InfluxDBServerDB +
		" WHERE host = '" + serverName + "' AND operation = 'measurement' ORDER BY time DESC LIMIT " + strconv.FormatInt(config.ThroughputMovingAverageInterval, 10)
	res, err := databases.QueryDB(config.InfluxDBConnection, query, config.InfluxDBServerDB)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
	for _, re := range res {
		if re.Series == nil {
			place := logger.Place()
			logger.Debug(place, "database throughput is nil")
			return 0
		}
	}

	total := 0.0
	for _, row := range res[0].Series[0].Values {
		throughput, err := row[1].(json.Number).Float64()
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}
		total += throughput
	}

	return total
}

// movingAverageOfThroughputRatio は 上限スループットupperLimitに対するその時点でのスループットの割合 の移動平均を算出します
// upperLimitは各VMのUpperLimit, 移動平均の区間は config.ThroughputMovingAverageInterval で指定します
func movingAverageOfThroughputRatio(serverName string, upperLimit float64, config *models.Config) float64 {
	query := "SELECT time, throughput FROM " + config.InfluxDBServerDB +
		" WHERE host = '" + serverName + "' AND operation = 'measurement' ORDER BY time DESC LIMIT " + strconv.FormatInt(config.ThroughputMovingAverageInterval, 10)
	res, err := databases.QueryDB(config.InfluxDBConnection, query, config.InfluxDBServerDB)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
	for _, re := range res {
		if re.Series == nil {
			place := logger.Place()
			logger.Debug(place, "database throughput is nil")
			return 0
		}
	}

	// 上限スループットupperLimitに対するその時点でのthroughputの割合 を
	// intervalで指定した区間分合計したもの
	totalRatioInInterval := 0.0
	interval := 0
	for i, row := range res[0].Series[0].Values {
		throughput, err := row[1].(json.Number).Float64()
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}
		totalRatioInInterval += throughput / upperLimit
		interval = i
	}

	movingAverage := totalRatioInInterval / float64(interval+1)
	tags := []string{"parameter:working_log", "operation:load_determination", "host:" + serverName}
	fields := []string{
		fmt.Sprintf("moving_average:%f", movingAverage),
	}
	logger.Record(tags, fields)
	databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

	return movingAverage
}

// スループットを用いて各サーバの負荷状況を判断する
// 0:普通 1:高負荷 2:低負荷
func judgeEachStatus(serverName string, average float64, config *models.Config) int {
	var val float64
	var twts [30]models.ThroughputWithTime

	query := "SELECT time, throughput FROM " + config.InfluxDBServerDB + " WHERE host = '" + serverName + "' AND operation = 'measurement' ORDER BY time DESC LIMIT 30"
	res, err := databases.QueryDB(config.InfluxDBConnection, query, config.InfluxDBServerDB)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}

	for _, re := range res {
		if re.Series == nil {
			place := logger.Place()
			logger.Debug(place, "database throughput is nil")
			return 0
		}
	}
	for i, row := range res[0].Series[0].Values {
		t, err := time.Parse(time.RFC3339, row[0].(string))
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}
		val, err = row[1].(json.Number).Float64()
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
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
	TPHigh := config.Cluster.VirtualMachines[serverName].ThroughputUpperLimit
	c := 0
	for i, twt := range twts {
		if twt.Throughput == 0 {
			break // これ以上データが無いため
		}
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
	TPHigh := config.Cluster.VirtualMachines[serverName].ThroughputUpperLimit
	ratio := config.ThroughputScaleInRatio
	c := 0
	for i, twt := range twts {
		if twt.Throughput == 0 {
			break // これ以上データが無いため
		}
		if twt.Throughput > float64(TPHigh)*ratio {
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
