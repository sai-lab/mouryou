package functions

import (
	"container/ring"

	"github.com/sai-lab/mouryou/lib/average"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/ratio"
)

func ServerManagement(config *models.ConfigStruct) {
	var avgor, out, in, high, low, ir, n float64
	var o, w, i int

	r := ring.New(LING_SIZE)
	avgors := make([]float64, LING_SIZE)

	for avgor = range loadCh {
		r.Value = avgor
		r = r.Next()

		o = mutex.Read(&operating, &operateMutex)
		if o > 0 {
			continue
		}

		avgors = convert.RingToArray(r)
		out = average.MovingAverage(avgors, config.Cluster.LoadBalancer.ScaleOut)
		in = average.MovingAverage(avgors, config.Cluster.LoadBalancer.ScaleIn)

		w = mutex.Read(&working, &workMutex)
		high = config.Cluster.LoadBalancer.ThHigh(w, len(config.Cluster.VirtualMachines))
		low = config.Cluster.LoadBalancer.ThLow(w)

		ir = ratio.Increase(avgors, config.Cluster.LoadBalancer.ScaleOut)
		n = ((ir*float64(config.Sleep)+out)/high - 1.0) * float64(w)

		switch {
		case w < len(config.Cluster.VirtualMachines) && int(n) > 0:
			for i = 0; i < int(n); i++ {
				if w+i < len(config.Cluster.VirtualMachines) {
					go config.Cluster.VirtualMachines[w+i].Bootup(config.Sleep, powerCh)
				}
			}
		case w > 1 && in < low && mutex.Read(&waiting, &waitMutex) == 0:
			go config.Cluster.VirtualMachines[w-1].Shutdown(config.Sleep, powerCh)
		}
	}
}
