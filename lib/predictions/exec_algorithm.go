package predictions

import "github.com/sai-lab/mouryou/lib/models"

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
	}
	return n, scaleIn
}

//ExecDifferentAlgorithm
func ExecDifferentAlgorithm(c *models.Config, w int, b int, s int, tw int, ttlORs []float64) int {
	var nw int // Necessary Weight
	switch c.Algorithm {
	case "PeriodicallyUseARMA":
		// 長期間のログデータを使ったARMAモデルに基づくオートスケールアルゴリズム
		nw = PeriodicallyPrediction(w, b, s, tw)
	case "ServerNumDependSpike":
		// 台数依存オートスケールアルゴリズム
		//config.Cluster.LoadBalancer.ChangeThresholdOut(w, b, s, len(config.Cluster.VirtualMachines))
	}
	return nw
}
