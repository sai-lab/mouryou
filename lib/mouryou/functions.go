package mouryou

import (
	"../math"
	"fmt"
	"log"
	"time"
)

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
	for avgor := range avgorCh {
		w := readWorking()
		thHigh := c.LB.thHigh()
		thLow := c.LB.thLow(w)

		switch {
		case w < len(c.VMs) && avgor > thHigh:
			powerCh <- "start"
		case w > 1 && avgor > thLow:
			powerCh <- "shutdown"
		}
	}
}

func DestinationSettingFunctin(c cluster) {
	for power := range powerCh {
		switch power {
		case "start":
			w := readWorking()
			c.LB.active(c.VMs[w].Host)
			writeWorking(w + 1)
		case "shutdown":
			w := readWorking()
			writeWorking(w - 1)
			c.LB.inactive(c.VMs[w-1].Host)
		}
	}
}
