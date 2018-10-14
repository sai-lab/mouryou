package engine

import (
	"fmt"
	"strings"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/timer"
)

func DestinationSetting(config *models.Config) {
	var b, s, w, o int

	connection, err := config.WebSocket.Dial()

	for power := range monitor.PowerCh {
		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shutting, &shutMutex)
		workingLog := []string{"workingLog", fmt.Sprintf("%d %d %d", w, b, s)}
		logger.Write(workingLog)

		if config.DevelopLogLevel >= 1 {
			fmt.Println("PowerCh comming ", power.Name, power.Info)
		}

		switch power.Info {
		case "booting up":
			mutex.Write(&booting, &bootMutex, b+1)
			logger.Send(connection, err, "Booting up: "+power.Name)
		case "booted up":
			config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[power.Name].Name)
			logger.Send(connection, err, "Booted up: "+power.Name)
			mutex.Write(&working, &workMutex, w+1)
			mutex.Write(&booting, &bootMutex, b-1)
			go timer.Set(&waiting, &waitMutex, config.Wait)
		case "shutting down":
			mutex.Write(&shutting, &shutMutex, s+1)
			mutex.Write(&working, &workMutex, w-1)
			config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[power.Name].Name)
			go timer.Set(&waiting, &waitMutex, config.Wait)
			logger.Send(connection, err, "Shutting down: "+power.Name)
		case "shutted down":
			mutex.Write(&shutting, &shutMutex, s-1)
			logger.Send(connection, err, "Shutted down: "+power.Name)
		default:
			fmt.Println("Error:", power)
			switch {
			case strings.Index(power.Info, "domain is already running") != -1:
				mutex.Write(&booting, &bootMutex, o-1)
			case strings.Index(power.Info, "domain is not running") != -1:
				mutex.Write(&shutting, &shutMutex, s-1)
			}
		}
	}
}
