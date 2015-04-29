package mouryou

import (
	"../math"
	"container/ring"
	"fmt"
	"log"
	"time"
)

const MaxLingSize = 10

var avgorCh = make(chan float64, 1)
var powerCh = make(chan string, 1)

func LoadMonitoringFunction(c cluster) {
	for {
		w := readWorking()
		ors := c.operatingRatios(w)

		str := sliceToCsv(ors)
		fmt.Println(str)
		log.Println(str)

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
			powerCh <- "create"
		case w > 1 && inAvgor < thLow:
			powerCh <- "shutdown"
		}
	}
}

func DestinationSettingFunctin(c cluster) {
	for power := range powerCh {
		switch power {
		case "create":
			w := readWorking()
			go c.VMs[w].create(30, true)
			writeOperating(true)
		case "created":
			w := readWorking()
			c.LB.active(c.VMs[w].Host)
			writeWorking(w + 1)
			writeOperating(false)
		case "shutdown":
			w := readWorking()
			writeWorking(w - 1)
			c.LB.inactive(c.VMs[w-1].Host)
			go c.VMs[w-1].shutdown(30, true)
			writeOperating(true)
		case "shutdowned":
			writeOperating(false)
		}
	}
}
