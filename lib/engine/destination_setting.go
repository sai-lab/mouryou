package engine

import (
	"fmt"
	"strings"

	//"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/timer"
)

func DestinationSetting(config *models.ConfigStruct) {
	var power PowerStruct
	var b, s, w, o int

	// connection, err := config.WebSocket.Dial()

	for power = range PowerCh {
		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shuting, &shutMutex)

		switch power.Info {
		case "booting up":
			mutex.Write(&booting, &bootMutex, b+1)
			// logger.Send(connection, err, "Booting up: "+strconv.Itoa(w))
		case "booted up":
			config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[power.Name].Name)
			// logger.Send(connection, err, "Booted up: "+strconv.Itoa(w))
			mutex.Write(&working, &workMutex, w+1)
			mutex.Write(&booting, &bootMutex, b-1)
			go timer.Set(&waiting, &waitMutex, config.Wait)
		case "shutting down":
			mutex.Write(&shuting, &shutMutex, s+1)
			mutex.Write(&working, &workMutex, w-1)
			// config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[power.Name].Name)
			// logger.Send(connection, err, "Shutting down: "+strconv.Itoa(w-1))
		case "shutted down":
			mutex.Write(&shuting, &shutMutex, s-1)
			// logger.Send(connection, err, "Shutted down: "+strconv.Itoa(w-1))
		default:
			fmt.Println("Error:", power)
			switch {
			case strings.Index(power.Info, "domain is already running") != -1:
				mutex.Write(&booting, &bootMutex, o-1)
			case strings.Index(power.Info, "domain is not running") != -1:
				mutex.Write(&shuting, &shutMutex, s-1)
			}
		}
	}
}
