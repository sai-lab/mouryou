package functions

import (
	"strconv"

	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/timer"
	"golang.org/x/net/websocket"
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

			if err == nil {
				websocket.Message.Send(connection, "Booting up: "+strconv.Itoa(w))
			}
		case "booted up":
			config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[w].Host)
			if err == nil {
				websocket.Message.Send(connection, "Booted up: "+strconv.Itoa(w))
			}

			mutex.Write(&working, &workMutex, w+1)
			mutex.Write(&operating, &operateMutex, o-1)

			go timer.Set(&waiting, &waitMutex, config.Sleep)
		case "shutting down":
			mutex.Write(&operating, &operateMutex, o+1)
			mutex.Write(&working, &workMutex, w-1)

			config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[w-1].Host)
			if err == nil {
				websocket.Message.Send(connection, "Shutting down: "+strconv.Itoa(w-1))
			}
		case "shutted down":
			mutex.Write(&operating, &operateMutex, o-1)

			if err == nil {
				websocket.Message.Send(connection, "Shutted down: "+strconv.Itoa(w-1))
			}
		}
	}
}
