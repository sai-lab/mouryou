package predictions

import "github.com/sai-lab/mouryou/lib/models"

//ExecDifferentAlgorithm exec algorithm for server with same performance
func ExecDifferentAlgorithm(c *models.Config, w int, b int, s int, tw int, ttlORs []float64) int {
	var nw int // Necessary Weight
	switch c.Algorithm {
	case "PeriodicallyUseARMA":
		// 長期間のログデータを使ったARMAモデルに基づくオートスケールアルゴリズム
		nw = PeriodicallyPrediction(w, b, s, tw)
	}
	return nw
}

//ExecSameAlgorithm exec algorithm for server with different performance
func ExecSameAlgorithm(c *models.Config, w int, b int, s int, tw int, ttlORs []float64) (float64, bool) {
	var n float64
	var scaleIn bool
	switch c.Algorithm {
	case "BasicSpike":
		// 短期間の移動平均に基づくオートスケールアルゴリズム
		n, scaleIn = basicSpike(c, w, b, s, tw, ttlORs)
	case "ServerNumDependSpike":
		// 台数依存オートスケールアルゴリズム
		c.Cluster.LoadBalancer.ChangeThresholdOut(w, b, s, len(c.Cluster.VirtualMachines))
		n, scaleIn = basicSpike(c, w, b, s, tw, ttlORs)
	case "DecreaseWeightFromBasicSpike":
		// 過負荷となったサーバの重みを下げるオートスケールアルゴリズム
		n, scaleIn = basicSpike(c, w, b, s, tw, ttlORs)
	}
	return n, scaleIn
}
