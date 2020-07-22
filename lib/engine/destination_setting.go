package engine

import (
	"fmt"
	"strings"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"golang.org/x/net/websocket"
)

// DestinationSetting では，monitor.PowerChを常に受信し，
// 受信したチャネルに従って，振分先情報の追加・削除をロードバランサに通知します．
func DestinationSetting(config *models.Config) {
	var b, s, w, o int
	var connection *websocket.Conn
	var err error

	if config.UseWeb {
		connection, err = config.WebSocket.Dial()
	}

	for power := range monitor.PowerCh {
		// 稼働中の台数
		w = mutex.Read(&working, &workMutex)
		// 起動処理中の台数
		b = mutex.Read(&booting, &bootMutex)
		// 停止処理中の台数
		s = mutex.Read(&shutting, &shutMutex)

		// データベースとログに稼働状況を記録
		tags := []string{
			"parameter:working_log",
			"operation:power",
			fmt.Sprintf("host:%s", power.Name),
			fmt.Sprintf("load:%s", power.Load),
		}
		fields := []string{
			fmt.Sprintf("working:%d", w),
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
		case "booting up": // 起動処理を開始した
			// 起動処理中の台数を増加
			mutex.Write(&booting, &bootMutex, b+1)
			if config.UseWeb {
				logger.Send(connection, err, "Booting up: "+power.Name)
			}
		case "booted up": // 起動処理が完了した
			// ロードバランサの振分先にpower.Nameを追加
			err := config.Cluster.LoadBalancer.Active(config.Cluster.VirtualMachines[power.Name].Name)
			if err != nil {
				place := logger.Place()
				logger.Error(place, err)
			}
			if config.UseWeb {
				logger.Send(connection, err, "Booted up: "+power.Name)
			}
			// 稼働中の台数を増加
			mutex.Write(&working, &workMutex, w+1)
			// 起動処理が終わったので，起動処理中の台数を減少
			mutex.Write(&booting, &bootMutex, b-1)
			// 起動処理が完了した後，config.Wait秒間は停止処理を発火しないようにwaitingを設定
			//go timer.Set(&waiting, &waitMutex, config.Wait)
		case "shutting down": // 停止処理を開始した
			// 停止処理中の台数を増加
			mutex.Write(&shutting, &shutMutex, s+1)
			// 稼働中の台数を減少
			mutex.Write(&working, &workMutex, w-1)
			// ロードバランサの振分先からpower.Nameを削除
			err := config.Cluster.LoadBalancer.Inactive(config.Cluster.VirtualMachines[power.Name].Name)
			if err != nil {
				place := logger.Place()
				logger.Error(place, err)
			}
			// 停止処理を開始した後，config.Wait秒間は停止処理を発行しないようにwaitingを設定
			go timer.Set(&waiting, &waitMutex, config.Wait)
			if config.UseWeb {
				logger.Send(connection, err, "Shutting down: "+power.Name)
			}
		case "shutted down": // 停止処理が完了した
			// 停止処理が終わったので，停止処理中の台数を減少
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
