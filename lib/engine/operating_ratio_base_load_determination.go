package engine

import (
	"container/ring"
	"fmt"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/predictions"
)

func operatingRatioBase(config *models.Config) {
	var (
		// totalOR means the total value of the operating ratios of the working servers
		totalOR float64
		// necessaryWeights means the necessary weights
		necessaryWeights int
		// orders is auto autoScaleOrder order
		orders []autoScaleOrder
	)

	r := ring.New(LingSize)
	ttlORs := make([]float64, LingSize)

	for totalOR = range monitor.LoadORCh {
		r.Value = totalOR
		r = r.Next()
		ttlORs = convert.RingToArray(r)

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
			"base_load:or",
			"parameter:working_log",
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

		// 負荷判定アルゴリズム実行
		if config.UseHetero {
			necessaryWeights = predictions.ExecDifferentAlgorithm(config, lWorking,
				lBooting, lShutting, lTotalWeight, lFutureTotalWeight, ttlORs)
			switch {
			case necessaryWeights > lTotalWeight:
				orders = append(orders, autoScaleOrder{Handle: "ScaleOut",
					Weight: necessaryWeights - lFutureTotalWeight, Load: "OR"})
			case necessaryWeights < lTotalWeight:
				orders = append(orders, autoScaleOrder{Handle: "ScaleIn",
					Weight: necessaryWeights - lFutureTotalWeight, Load: "OR"})
			}
		} else {
			orders = scaleSameServers(config, ttlORs, lWorking, lBooting,
				lShutting, lTotalWeight, lFutureTotalWeight)
		}

		for _, order := range orders {
			autoScaleOrderCh <- order
		}
	}
}

// scaleSameServersは単一性能向けアルゴリズムのサーバ起動停止メソッドです.
// predictions.ExecSameAlgorithmメソッドからmodels.Config.Sleep時間後に必要な台数と
// スケールインするかの真偽値を受け取り,それらに従って起動停止処理を実行します.
func scaleSameServers(c *models.Config, ttlORs []float64, working int, booting int,
	shutting int, tw int, fw int) []autoScaleOrder {
	var (
		scaleIn        bool
		requiredNumber float64
		orders         []autoScaleOrder
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
			orders = append(orders, autoScaleOrder{Handle: "ScaleOut", Weight: 10, Load: "OR"})
		}
	case working > 1 && scaleIn && mutex.Read(&waiting, &waitMutex) == 0 && booting == 0:
		orders = append(orders, autoScaleOrder{Handle: "ScaleIn", Weight: 10, Load: "OR"})
	}

	return orders
}
