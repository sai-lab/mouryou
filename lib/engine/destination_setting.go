package engine

import (
	"fmt"
	"strings"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/timer"
	"golang.org/x/net/websocket"
)

func DestinationSetting(config *models.Config) {
	var b, s, w, o int
	var connection *websocket.Conn
	var err error

	if config.UseWeb {
		connection, err = config.WebSocket.Dial()
	}

	for power := range monitor.PowerCh {
		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shutting, &shutMutex)

		tags := []string{"parameter:working_log", "operation:power", fmt.Sprintf("host:%s", power.Name)}
		fields := []string{fmt.Sprintf("working:%d", w),
			fmt.Sprintf("booting:%d", b),
			fmt.Sprintf("shutting:%d", s),
			fmt.Sprintf("power:%s", power.Info),
		}
		logger.Record(tags, fields)
		databases.WriteValues(config.InfluxDBConnection, config, tags, fields)

		if config.DevelopLogLevel >= 1 {
			fmt.Println("PowerCh comming ", power.Name, power.Info)
		}

		switch power.Info {
		case "booting up":
			mutex.Write(&booting, &bootMutex, b+1)
			if config.UseWeb {
				logger.Send(connection, err, "Booting up: "+power.Name)
			}
		case "booted up":
			err := config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[power.Name].Name)
			if err != nil {
				place := logger.Place()
				logger.Error(place, err)
			}
			if config.UseWeb {
				logger.Send(connection, err, "Booted up: "+power.Name)
			}
			mutex.Write(&working, &workMutex, w+1)
			mutex.Write(&booting, &bootMutex, b-1)
			go timer.Set(&waiting, &waitMutex, config.Wait)
		case "shutting down":
			mutex.Write(&shutting, &shutMutex, s+1)
			mutex.Write(&working, &workMutex, w-1)
			err := config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[power.Name].Name)
			if err != nil {
				place := logger.Place()
				logger.Error(place, err)
			}
			go timer.Set(&waiting, &waitMutex, config.Wait)
			if config.UseWeb {
				logger.Send(connection, err, "Shutting down: "+power.Name)
			}
		case "shutted down":
			mutex.Write(&shutting, &shutMutex, s-1)
			if config.UseWeb {
				logger.Send(connection, err, "Shutted down: "+power.Name)
			}
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
