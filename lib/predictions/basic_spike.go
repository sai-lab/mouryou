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
	num := len(c.Cluster.VirtualMachines)

	ThHigh := c.Cluster.LoadBalancer.ThHigh(c, w, num)
	ThLow := c.Cluster.LoadBalancer.ThLow(c, w, num)

	ir := ratio.Increase(ttlORs, c.Cluster.LoadBalancer.ScaleOut)
	predictedValue := out + ir*float64(c.Sleep)
	n := (predictedValue / ThHigh) - float64(w+b)
	weights := []string{"we", fmt.Sprintf("%3.5f", n), fmt.Sprintf("%3d", tw), fmt.Sprintf("%3d", fw), fmt.Sprintf("%3.5f", predictedValue)}
	logger.Print(weights)
	logger.Write(weights)
	scaleInLog := []string{"scaleInLog", fmt.Sprintf("%3.5f %3.5f %b %d %d %f", in, ThLow, in < ThLow, w, num, models.Threshold)}
	logger.Write(scaleInLog)

	return n, in < ThLow
}
