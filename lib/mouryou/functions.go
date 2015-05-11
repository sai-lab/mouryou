package mouryou

import (
	"container/ring"
	"time"

	"../average"
	"../rate"
)

const MaxLingSize = 10

var avgorCh = make(chan float64, 1)
var powerCh = make(chan string, 1)

func LoadMonitoringFunction(c cluster) {
	for {
		w := readWorking()
		ors := c.operatingRatios(w)
		logging(ors)

		avgorCh <- average.Average(ors)
		time.Sleep(time.Second)
	}
}

func ServerManagementFunctin(c cluster) {
	r := ring.New(MaxLingSize)

	for avgor := range avgorCh {
		if readOperating() > 0 {
			continue
		}

		r.Value = avgor
		r = r.Next()
		avgors := rtoa(r)

		outAvgor := average.MovingAverage(avgors, c.LB.ScaleOut)
		inAvgor := average.MovingAverage(avgors, c.LB.ScaleIn)

		w := readWorking()
		thHigh := c.LB.thHigh()
		thLow := c.LB.thLow(w)

		switch {
		case w < len(c.VMs) && outAvgor > thHigh:
			ri := rate.Increase(avgors)
			writeOperating(ri)

			for i := 0; i < ri; i++ {
				go c.VMs[w+i].create(wait)
			}
		case w > 1 && inAvgor < thLow:
			powerCh <- "shutdowning"
			writeOperating(1)
			go c.VMs[w-1].shutdown(wait)
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
