package functions

import (
	"container/ring"

	"github.com/sai-lab/mouryou/lib/average"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/ratio"
)

func ServerManagement(cluster *models.ClusterStruct) {
	var avgor, in, high, low, n float64
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
		in = average.MovingAverage(avgors, cluster.LoadBalancer.ScaleIn)

		w = mutex.Read(&working, &workMutex)
		high = cluster.LoadBalancer.ThHigh()
		low = cluster.LoadBalancer.ThLow(w)

		n = (ratio.Increase(avgors)*float64(SLEEP_SEC)+avgors[len(avgors)-1])/high - float64(o-1)

		switch {
		case w < len(cluster.VirtualMachines) && int(n) > 0:
			for i = 0; i < int(n); i++ {
				go cluster.VirtualMachines[w+i].Bootup(SLEEP_SEC, powerCh)
			}
		case w > 1 && in < low && mutex.Read(&working, &workMutex) != 1:
			go cluster.VirtualMachines[w-1].Shutdown(SLEEP_SEC, powerCh)
		}
	}
}
