package predictions

import "github.com/sai-lab/mouryou/lib/models"

func Exec(c *models.ConfigStruct, w int, b int, s int, tw int, ttlORs []float64) (float64, bool) {
	var n float64
	var scaleIn bool
	switch c.Algorithm {
	case "BasicSpike":
		// 短期間の移動平均に基づくオートスケールアルゴリズム
		n, scaleIn = basicSpike(c, w, b, s, tw, ttlORs)
	case "PeriodicallyUseARMA":
		// 長期間のログデータに基づくARMAモデルを使ったオートスケールアルゴリズム
	case "ServerNumDependSpike":
		// 台数依存オートスケールアルゴリズム
		//config.Cluster.LoadBalancer.ChangeThresholdOut(w, b, s, len(config.Cluster.VirtualMachines))
	}
	return n, scaleIn
}
