package predictions

import (
	"fmt"

	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/ratio"
)

// basicSpike
func basicSpike(c *models.Config, w int, b int, s int, tw int, fw int, ttlORs []float64) (float64, bool) {
	out := calculate.MovingAverage(ttlORs, c.Cluster.LoadBalancer.ScaleOut)
	in := calculate.MovingAverage(ttlORs, c.Cluster.LoadBalancer.ScaleIn)

	ThHigh := c.Cluster.LoadBalancer.ThHigh(c, w, len(c.Cluster.VirtualMachines))
	ThLow := c.Cluster.LoadBalancer.ThLow(c, w)

	ir := ratio.Increase(ttlORs, c.Cluster.LoadBalancer.ScaleOut)
	n := (((out + ir*float64(c.Sleep)) / ThHigh) - float64(w+b)) * 10
	weights := []string{"we", fmt.Sprintf("%3.5f", n), fmt.Sprintf("%3d", tw), fmt.Sprintf("%3d", fw)}
	logger.Print(weights)
	logger.Write(weights)

	return n, in < ThLow
}
