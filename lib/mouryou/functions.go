package mouryou

import (
	"log"
	"time"
)

func LoadMonitoringFunction(c cluster) {
	for {
		ors := c.operatingRatios()
		log.Printf("%+v\n", ors)
		time.Sleep(time.Second)
	}
}
