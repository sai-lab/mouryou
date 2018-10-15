package predictions

import (
	"fmt"

	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/ratio"
)

// basicSpike
func basicSpike(config *models.Config, w int, b int, s int, tw int, fw int, ttlORs []float64) (float64, bool) {
	out := calculate.MovingAverage(ttlORs, config.Cluster.LoadBalancer.ScaleOut)
	in := calculate.MovingAverage(ttlORs, config.Cluster.LoadBalancer.ScaleIn)
	num := len(config.Cluster.VirtualMachines)

	ThHigh := config.Cluster.LoadBalancer.ThHigh(config, w, num)
	ThLow := config.Cluster.LoadBalancer.ThLow(config, w, num)

	ir := ratio.Increase(ttlORs, config.Cluster.LoadBalancer.ScaleOut)
	predictedValue := out + ir*float64(config.Sleep)
	n := (predictedValue / ThHigh) - float64(w+b)

	tags := []string{"base_load:or", "operation:predict"}
	fields := []string{fmt.Sprintf("sleep:%d", config.Sleep),
		fmt.Sprintf("predicted_value:%3.5f", predictedValue)}
	logger.Record(tags, fields)
	databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

	tags = []string{"base_load:or", "operation:scale_in"}
	fields = []string{fmt.Sprintf("scale_in_material:%3.5f", in),
		fmt.Sprintf("th_low:%3.5f", ThLow),
		fmt.Sprintf("need_scale_in:%t", in < ThLow),
		fmt.Sprintf("working:%d", w),
		fmt.Sprintf("vm_num:%d", num),
		fmt.Sprintf("threshold:%f", models.Threshold),
	}
	logger.Record(tags, fields)
	databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

	return n, in < ThLow
}
