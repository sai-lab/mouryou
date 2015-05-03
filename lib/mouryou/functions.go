package mouryou

import (
	"../math"
	"container/ring"
	"time"
)

const MaxLingSize = 10

var avgorCh = make(chan float64, 1)
var powerCh = make(chan string, 1)

func LoadMonitoringFunction(c cluster) {
	for {
		w := readWorking()
		ors := c.operatingRatios(w)
		logging(ors)

		avgorCh <- math.Average(ors)
		time.Sleep(time.Second)
	}
}

func ServerManagementFunctin(c cluster) {
	r := ring.New(MaxLingSize)

	for avgor := range avgorCh {
		if readOperating() {
			continue
		}

		r.Value = avgor
		r = r.Next()
		avgors := rtoa(r)

		outAvgor := math.MovingAverage(avgors, c.LB.ScaleOut)
		inAvgor := math.MovingAverage(avgors, c.LB.ScaleIn)

		w := readWorking()
		thHigh := c.LB.thHigh()
		thLow := c.LB.thLow(w)

		switch {
		case w < len(c.VMs) && outAvgor > thHigh:
			go c.VMs[w].create(wait)
			writeOperating(true)
		case w > 1 && inAvgor < thLow:
			powerCh <- "shutdowning"
			go c.VMs[w-1].shutdown(wait)
			writeOperating(true)
		}
	}
}

func DestinationSettingFunctin(c cluster) {
	for power := range powerCh {
		w := readWorking()
		switch power {
		case "created":
			c.LB.active(c.VMs[w].Host)
			writeWorking(w + 1)
		case "shutdowning":
			writeWorking(w - 1)
			c.LB.inactive(c.VMs[w-1].Host)
		}
	}
}
