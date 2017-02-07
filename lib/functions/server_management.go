package functions

import (
	"container/ring"

	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/ratio"
)

func ServerManagement(config *models.ConfigStruct) {
	var ttlor, out, in, ThHigh, ThLow, ir, n float64
	var b, w, s, i int

	r := ring.New(LING_SIZE)
	ttlors := make([]float64, LING_SIZE)

	for ttlor = range loadCh {
		r.Value = ttlor
		r = r.Next()

		ttlors = convert.RingToArray(r)
		out = calculate.MovingAverage(ttlors, config.Cluster.LoadBalancer.ScaleOut)
		in = calculate.MovingAverage(ttlors, config.Cluster.LoadBalancer.ScaleIn)

		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shuting, &shutMutex)
		// config.Cluster.LoadBalancer.ChangeThresholdOut(w, len(config.Cluster.VirtualMachines))
		ThHigh = config.Cluster.LoadBalancer.ThHigh(w, len(config.Cluster.VirtualMachines))
		ThLow = config.Cluster.LoadBalancer.ThLow(w)

		ir = ratio.Increase(ttlors, config.Cluster.LoadBalancer.ScaleOut)
		n = ((out + ir*float64(config.Sleep)) / ThHigh) - float64(w+b)

		switch {
		case w+b < len(config.Cluster.VirtualMachines) && int(n) > 0 && s == 0:
			for i = 0; i < int(n); i++ {
				if w+b+i < len(config.Cluster.VirtualMachines) {
					go config.Cluster.VirtualMachines[w+b+i].Bootup(config.Sleep, powerCh)
					logger.PrintPlace("Bootup")
				}
			}
		case w > 1 && in < ThLow && mutex.Read(&waiting, &waitMutex) == 0 && b == 0:
			go config.Cluster.VirtualMachines[w-1].Shutdown(config.Sleep, powerCh)
			logger.PrintPlace("Shutdown")
		}
	}
}
