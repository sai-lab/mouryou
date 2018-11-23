package predictions

import (
	"fmt"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

//ExecDifferentAlgorithm exec algorithm for server with same performance
func ExecDifferentAlgorithm(config *models.Config, w int, b int, s int, tw int, fw int, ttlORs []float64) int {
	var nw int // Necessary Weight
	switch config.Cluster.LoadBalancer.OperatingRatioAlgorithm {
	case "PeriodicallyUseARMA":
		// 長期間のログデータを使ったARMAモデルに基づくオートスケールアルゴリズム
		nw = PeriodicallyPrediction(w, b, s, tw, fw)
	}
	return nw
}

//ExecSameAlgorithm exec algorithm for server with different performance
func ExecSameAlgorithm(config *models.Config, w int, b int, s int, tw int, fw int, ttlORs []float64) (float64, bool) {
	var n float64
	var scaleIn bool
	switch config.Cluster.LoadBalancer.OperatingRatioAlgorithm {
	case "BasicSpike":
		// 短期間の移動平均に基づくオートスケールアルゴリズム
		n, scaleIn = basicSpike(config, w, b, s, tw, fw, ttlORs)
	case "ServerNumDependSpike":
		// 台数依存オートスケールアルゴリズム
		changedThreshold, operatingUnitRatio := config.Cluster.LoadBalancer.ChangeThresholdOutInOperatingRatioAlgorithm(w, b, len(config.Cluster.VirtualMachines))
		loggingThreshold(config, changedThreshold, operatingUnitRatio, w, b, s)
		n, scaleIn = basicSpike(config, w, b, s, tw, fw, ttlORs)
	case "DecreaseWeightFromBasicSpike":
		// 過負荷となったサーバの重みを下げるオートスケールアルゴリズム
		n, scaleIn = basicSpike(config, w, b, s, tw, fw, ttlORs)
	}

	tags := []string{"base_load:or"}
	fields := []string{fmt.Sprintf("total_weight:%d", tw),
		fmt.Sprintf("future_total_weight:%d", fw),
		fmt.Sprintf("required_num:%3.5f", n),
	}
	logger.Record(tags, fields)
	databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

	return n, scaleIn
}

func loggingThreshold(config *models.Config, thresholdOut float64, operatingUnitRatio, work, boot, shut int) {
	tags := []string{
		"base_load:or",
		"operation:operating_ratio_base_load_determination",
		"parameter:threshold_out_log",
	}
	fields := []string{
		fmt.Sprintf("threshold_out:%f", thresholdOut),
		fmt.Sprintf("operating_unit_ratio:%d", operatingUnitRatio),
		fmt.Sprintf("working:%d", work),
		fmt.Sprintf("booting:%d", boot),
		fmt.Sprintf("shutting:%d", shut),
	}
	logger.Record(tags, fields)
	databases.WriteValues(config.InfluxDBConnection, config, tags, fields)
}
