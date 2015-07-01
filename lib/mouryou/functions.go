package mouryou

import (
	"container/ring"
	"time"

	"../average"
	"../ratio"
)

const (
	LING_SIZE   = 10
	TIMEOUT_SEC = 1
	SLEEP_SEC   = 30
)

var loadCh = make(chan float64, 1)
var powerCh = make(chan string, 1)

func LoadMonitoringFunction(cluster *ClusterStruct) {
	var n int

	for {
		n = readWorking()
		ors := cluster.OperatingRatios(n)
		logging(ors)

		loadCh <- average.Average(ors)
		time.Sleep(time.Second)
	}
}

func ServerManagementFunctin(cluster *ClusterStruct) {
	r := ring.New(LING_SIZE)

	for avgor := range loadCh {
		if readOperating() > 0 {
			continue
		}

		r.Value = avgor
		r = r.Next()
		avgors := rtoa(r)

		out := average.MovingAverage(avgors, cluster.LoadBalancer.ScaleOut)
		in := average.MovingAverage(avgors, cluster.LoadBalancer.ScaleIn)

		n := readWorking()
		high := cluster.LoadBalancer.ThHigh()
		low := cluster.LoadBalancer.ThLow(n)

		switch {
		case n < len(cluster.VirtualMachines) && out > high:
			ir := ratio.Increase(avgors)
			writeOperating(ir)

			for i := 0; i < ir; i++ {
				go cluster.VirtualMachines[n+i].Bootup(SLEEP_SEC)
			}
		case n > 1 && in < low:
			writeOperating(1)
			go cluster.VirtualMachines[n-1].Shutdown(SLEEP_SEC)
		}
	}
}

func DestinationSettingFunctin(cluster *ClusterStruct) {
	for power := range powerCh {
		n := readWorking()

		switch power {
		case "bootup":
			cluster.LoadBalancer.Active(cluster.VirtualMachines[n].Host)
			writeWorking(n + 1)
		case "shutdown":
			writeWorking(n - 1)
			cluster.LoadBalancer.Inactive(cluster.VirtualMachines[n].Host)
		}
	}
}
