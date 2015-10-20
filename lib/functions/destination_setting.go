package functions

import (
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/timer"
)

func DestinationSetting(config *models.ConfigStruct) {
	var power string
	var w, o int

	for power = range powerCh {
		w = mutex.Read(&working, &workMutex)
		o = mutex.Read(&operating, &operateMutex)

		switch power {
		case "booting up":
			mutex.Write(&operating, &operateMutex, o+1)
		case "booted up":
			config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[w].Host)
			mutex.Write(&working, &workMutex, w+1)
			mutex.Write(&operating, &operateMutex, o-1)
			go timer.Set(&waiting, &waitMutex, config.Sleep)
		case "shutting down":
			mutex.Write(&operating, &operateMutex, o+1)
			mutex.Write(&working, &workMutex, w-1)
			config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[w-1].Host)
		case "shutted down":
			mutex.Write(&operating, &operateMutex, o-1)
		}
	}
}
