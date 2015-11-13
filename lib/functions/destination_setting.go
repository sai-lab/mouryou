package functions

import (
	"strconv"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/timer"
)

func DestinationSetting(config *models.ConfigStruct) {
	var power string
	var w, o int

	connection, err := config.WebSocket.Dial()

	for power = range powerCh {
		w = mutex.Read(&working, &workMutex)
		o = mutex.Read(&operating, &operateMutex)

		switch power {
		case "booting up":
			mutex.Write(&operating, &operateMutex, o+1)
			logger.Send(connection, err, "Booting up: "+strconv.Itoa(w))
		case "booted up":
			config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[w].Host)
			logger.Send(connection, err, "Booted up: "+strconv.Itoa(w))
			mutex.Write(&working, &workMutex, w+1)
			mutex.Write(&operating, &operateMutex, o-1)
			go timer.Set(&waiting, &waitMutex, config.Wait)
		case "shutting down":
			mutex.Write(&operating, &operateMutex, o+1)
			mutex.Write(&working, &workMutex, w-1)
			config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[w-1].Host)
			logger.Send(connection, err, "Shutting down: "+strconv.Itoa(w-1))
		case "shutted down":
			mutex.Write(&operating, &operateMutex, o-1)
			logger.Send(connection, err, "Shutted down: "+strconv.Itoa(w-1))
		}
	}
}
