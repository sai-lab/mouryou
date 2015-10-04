package functions

import (
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func DestinationSetting(cluster *models.ClusterStruct) {
	var power string
	var w, o int

	for power = range powerCh {
		w = mutex.Read(&working, &workMutex)
		o = mutex.Read(&operating, &operateMutex)

		switch power {
		case "booting up":
			mutex.Write(&operating, o+1, &operateMutex)
		case "booted up":
			cluster.LoadBalancer.Active(cluster.VirtualMachines[w].Host)
			mutex.Write(&working, w+1, &workMutex)
			mutex.Write(&operating, o-1, &operateMutex)
		case "shutting down":
			mutex.Write(&operating, o+1, &operateMutex)
			mutex.Write(&working, w-1, &workMutex)
			cluster.LoadBalancer.Inactive(cluster.VirtualMachines[w-1].Host)
		case "shutted down":
			mutex.Write(&operating, o-1, &operateMutex)
		}
	}
}
