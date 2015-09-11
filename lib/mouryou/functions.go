package mouryou

import (
	"container/ring"
	"net/http"
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/average"
	"github.com/sai-lab/mouryou/lib/ratio"
)

const (
	LING_SIZE   = 10
	TIMEOUT_SEC = 1
	SLEEP_SEC   = 30
)

var (
	loadCh           = make(chan float64, 1)
	powerCh          = make(chan string, 1)
	working      int = 1
	operating    int = 0
	workMutex    sync.RWMutex
	operateMutex sync.RWMutex
)

func LoadMonitoringFunction(cluster *ClusterStruct) {
	var w int

	http.DefaultClient.Timeout = time.Duration(TIMEOUT_SEC * time.Second)

	for {
		w = readWithMutex(&working, &workMutex)
		ors := cluster.OperatingRatios(w)
		logging(ors)

		loadCh <- average.Average(ors)
		time.Sleep(time.Second)
	}
}

func ServerManagementFunctin(cluster *ClusterStruct) {
	var avgor, out, in, high, low float64
	var w, i, ir int

	r := ring.New(LING_SIZE)

	for avgor = range loadCh {
		if readWithMutex(&operating, &operateMutex) > 0 {
			continue
		}

		r.Value = avgor
		r = r.Next()
		avgors := RingToArray(r)

		out = average.MovingAverage(avgors, cluster.LoadBalancer.ScaleOut)
		in = average.MovingAverage(avgors, cluster.LoadBalancer.ScaleIn)

		w = readWithMutex(&working, &workMutex)
		high = cluster.LoadBalancer.ThHigh()
		low = cluster.LoadBalancer.ThLow(w)

		switch {
		case w < len(cluster.VirtualMachines) && out > high:
			ir = ratio.Increase(avgors)

			for i = 0; i < ir; i++ {
				go cluster.VirtualMachines[w+i].Bootup(SLEEP_SEC, powerCh)
			}
		case w > 1 && in < low:
			go cluster.VirtualMachines[w-1].Shutdown(SLEEP_SEC, powerCh)
		}
	}
}

func DestinationSettingFunctin(cluster *ClusterStruct) {
	var power string
	var w, o int

	for power = range powerCh {
		w = readWithMutex(&working, &workMutex)
		o = readWithMutex(&operating, &operateMutex)

		switch power {
		case "booting up":
			writeWithMutex(&operating, o+1, &operateMutex)
		case "booted up":
			cluster.LoadBalancer.Active(cluster.VirtualMachines[w].Host)
			writeWithMutex(&working, w+1, &workMutex)
			writeWithMutex(&operating, o-1, &operateMutex)
		case "shutting down":
			writeWithMutex(&operating, o+1, &operateMutex)
			writeWithMutex(&working, w-1, &workMutex)
			cluster.LoadBalancer.Inactive(cluster.VirtualMachines[w-1].Host)
		case "shutted down":
			writeWithMutex(&operating, o-1, &operateMutex)
		}
	}
}
