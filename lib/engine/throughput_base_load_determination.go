package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func throughputBase(config *models.Config) {
	for _ = range monitor.LoadTPCh {
		// スケールアウトフラグ
		shouldScaleOut := false
		// スケールインフラグ
		shouldScaleIn := false

		// 稼働中のサーバ名を格納
		var bootedServersName []string
		// 起動処理中のサーバ名を格納
		var bootingServersName []string

		// 稼働中の台数
		lWorking := mutex.Read(&working, &workMutex)
		// 起動処理中の台数
		lBooting := mutex.Read(&booting, &bootMutex)
		// 停止処理中の台数
		lShutting := mutex.Read(&shutting, &shutMutex)
		// 稼働中サーバの重みの合計値
		lTotalWeight := mutex.Read(&totalWeight, &totalWeightMutex)
		// サーバの起動・停止処理完了後の重みの合計値
		lFutureTotalWeight := mutex.Read(&futureTotalWeight, &futureTotalWeightMutex)

		// データベースとログに稼働状況を記録
		tags := []string{
			"base_load:th",
			"parameter:working_log",
			"operation:throughput_base_load_determination",
		}
		fields := []string{
			fmt.Sprintf("working:%d", lWorking),
			fmt.Sprintf("booting:%d", lBooting),
			fmt.Sprintf("shutting:%d", lShutting),
			fmt.Sprintf("total_weight:%d", lTotalWeight),
			fmt.Sprintf("future_total_weight:%d", lFutureTotalWeight),
		}
		logger.Record(tags, fields)
		databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

		// 稼働中のサーバ名や起動処理中のサーバ名を取得
		for _, v := range monitor.GetServerStates() {
			if config.DevelopLogLevel >= 6 {
				place := logger.Place()
				logger.Debug(place, "GetServerStates() Machine Name: "+v.Name+"Machine Info: "+v.Info)
			}
			if v.Info == "booted up" {
				bootedServersName = append(bootedServersName, v.Name)
			} else if v.Info == "booting up" {
				bootingServersName = append(bootingServersName, v.Name)
			}
		}

		switch config.ThroughputAlgorithm {
		case "MovingAverageV1.2":
			// 稼働中，起動処理中のサーバを一まとまりのクラスタと仮定して移動平均を計算する負荷判定機能
			// 20181024以降利用
			shouldScaleOut, shouldScaleIn = judgeByMovingAverageForCluster(config, bootedServersName, bootingServersName)
		case "All":
			// 全サーバが同時に過負荷や低負荷になるとオートスケールが必要と判断する負荷判定機能
			// 20181020以降未使用，比較実験用に残す
			shouldScaleOut, shouldScaleIn = judgeByAllServerSameLoad(config, bootedServersName)
		case "MovingAverage":
			// サーバ毎にスループットの上限スループットに対する割合の移動平均を計算してから，全サーバの平均を計算する負荷判定機能
			// 起動処理中のサーバの重みを考慮しづらいため廃止(20181020考案，20181023廃止)，比較実験用に残す
			shouldScaleOut, shouldScaleIn = judgeByMovingAverageForEachServer(config, bootedServersName)
		default:
			panic("unknown algorithm")
		}
		if shouldScaleOut {
			autoScaleOrderCh <- autoScaleOrder{Handle: "ScaleOut", Weight: 10, Load: "TP"}
		} else if shouldScaleIn {
			autoScaleOrderCh <- autoScaleOrder{Handle: "ScaleIn", Weight: 10, Load: "TP"}
		}
	}
}

// 稼働中，起動処理中のサーバを一まとまりのクラスタと仮定して移動平均を計算する負荷判定機能
// 20181024以降利用
func judgeByMovingAverageForCluster(config *models.Config, bootedServersName []string, bootingServersName []string) (bool, bool) {
	// スケールアウトフラグ
	shouldScaleOut := false
	// スケールインフラグ
	shouldScaleIn := false

	totalThroughput := 0.0
	totalUpperLimit := 0.0
	for _, name := range bootedServersName {
		value := config.Cluster.VirtualMachines[name]
		totalThroughput += intervalThroughputTotal(name, config)
		totalUpperLimit += value.ThroughputUpperLimit
	}
	for _, name := range bootingServersName {
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
		shouldScaleOut = true
	} else if movingAverage <= config.ThroughputScaleInRatio {
		shouldScaleIn = true
	}

	return shouldScaleOut, shouldScaleIn
}

func intervalThroughputTotal(serverName string, config *models.Config) float64 {
	query := "SELECT time, throughput FROM " +
		config.InfluxDBServerDB +
		" WHERE host = '" +
		serverName +
		"' AND operation = 'measurement' ORDER BY time DESC LIMIT " +
		strconv.FormatInt(config.ThroughputMovingAverageInterval, 10)
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

// サーバ毎にスループットの上限スループットに対する割合の移動平均を計算してから
// 全サーバの平均を計算する負荷判定機能
// 起動処理中のサーバの重みを考慮しづらいため廃止(20181020考案，20181023廃止)
func judgeByMovingAverageForEachServer(config *models.Config, bootedServersName []string) (bool, bool) {
	// スケールアウトフラグ
	shouldScaleOut := false
	// スケールインフラグ
	shouldScaleIn := false

	totalTPRatioMovingAverage := 0.0
	for _, name := range bootedServersName {
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
	ratioAverage := totalTPRatioMovingAverage / float64(len(bootedServersName))

	// ログにratioAverageを記録
	tags := []string{"parameter:working_log", "operation:load_determination"}
	fields := []string{
		fmt.Sprintf("ratio_average:%f", ratioAverage),
	}
	logger.Record(tags, fields)
	databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

	// 判定
	if ratioAverage >= config.ThroughputScaleOutRatio {
		shouldScaleOut = true
	} else if ratioAverage <= config.ThroughputScaleInRatio {
		shouldScaleIn = true
	}

	return shouldScaleOut, shouldScaleIn
}

// movingAverageOfThroughputRatio は 上限スループットupperLimitに対するその時点でのスループットの割合 の移動平均を算出します
// upperLimitは各VMのUpperLimit, 移動平均の区間は config.ThroughputMovingAverageInterval で指定します
func movingAverageOfThroughputRatio(serverName string, upperLimit float64, config *models.Config) float64 {
	query := "SELECT time, throughput FROM " +
		config.InfluxDBServerDB +
		" WHERE host = '" +
		serverName +
		"' AND operation = 'measurement' ORDER BY time DESC LIMIT " +
		strconv.FormatInt(config.ThroughputMovingAverageInterval, 10)
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

// 全てのサーバが上限スループットを超えるスループットを記録したならばスケールアウトが必要，
// 全てのサーバが上限スループットのR割以下のスループットを記録したならばスケールインが必要
// と判断する負荷判定方法
// 20180922に考案したが，振分量は常に均一ではないため，判定基準として正しくない
// 20181024からMovingAverageV1.2をメインで利用
// 比較実験用にコードは残す
func judgeByAllServerSameLoad(config *models.Config, bootedServersName []string) (bool, bool) {
	// スケールアウトフラグ
	shouldScaleOut := false
	// スケールインフラグ
	shouldScaleIn := false

	for _, name := range bootedServersName {
		value := config.Cluster.VirtualMachines[name]

		value.LoadStatus = judgeEachStatus(name, value.ThroughputUpperLimit, config)
		// 0:普通 1:高負荷 2:低負荷
		switch value.LoadStatus {
		case 1:
			shouldScaleOut = true
		case 2:
			shouldScaleIn = true
		default:
			shouldScaleIn = false
			shouldScaleOut = false
		}
	}
	return shouldScaleOut, shouldScaleIn
}

// スループットを用いて各サーバの負荷状況を判断する
// judgeByAllServerSameLoad で使用
// 0:普通 1:高負荷 2:低負荷
func judgeEachStatus(serverName string, average float64, config *models.Config) int {
	var val float64
	var twts [30]models.ThroughputWithTime

	query := "SELECT time, throughput FROM " +
		config.InfluxDBServerDB +
		" WHERE host = '" +
		serverName +
		"' AND operation = 'measurement' ORDER BY time DESC LIMIT 30"
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
// judgeByAllServerSameLoad 内で使用
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
// judgeByAllServerSameLoad 内で使用
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
