package monitor

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"golang.org/x/net/websocket"
)

func LoadMonitoring(config *models.Config) {
	var connection *websocket.Conn
	var err error
	http.DefaultClient.Timeout = time.Duration(config.Timeout * time.Second)
	if config.UseWeb {
		connection, err = config.WebSocket.Dial()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("ok")
		}
	}

	for {
		var bootedServers []string
		throughputs := make([]float64, len(config.Cluster.VirtualMachines))
		tw := 0
		//
		for _, v := range GetServerStates() {
			if config.DevelopLogLevel >= 6 {
				place := logger.Place()
				logger.Debug(place, "GetServerStates() Machine Name: "+v.Name+"Machine Info: "+v.Info)
			}
			if v.Info == "booted up" {
				bootedServers = append(bootedServers, v.Name)
				tw += v.Weight
			}
		}

		statuses := config.Cluster.ServerStatuses(bootedServers, config)
		sockets := config.Cluster.SocketStatuses(bootedServers, config)
		for i := range statuses {
			throughputs[i] = databases.WritePoints(config.InfluxDBConnection, config, statuses[i])
		}
		ors, arrays := Ratios(statuses, throughputs, tw, sockets)

		logger.PWArrays(config.DevelopLogLevel, arrays)
		if config.UseWeb {
			logger.Send(connection, err, convert.FloatsToStringsSimple(throughputs))
		}

		if config.UseOperatingRatio {
			LoadORCh <- calculate.Sum(ors)
		}
		if config.UseThroughput {
			LoadTPCh <- calculate.Sum(ors)
		}

		time.Sleep(time.Second)
	}
}

func Ratios(states []apache.ServerStatus, ths []float64, tw int, sockets []apache.SocketStatus) ([]float64, [13][]string) {
	var (
		operatingRatio    = 0
		cpuUsedPercent    = 1
		throughput        = 2 // throughput is ApacheLog in apache.ServerStatus
		dstatLog          = 3
		memoryUsedPersent = 4
		memoryBuffer      = 5
		memoryCached      = 6
		memoryStat        = 7
		acquisitionTime   = 8
		critical          = 9
		reqPerSec         = 10 // reqPerSecは起動してからの平均の1秒間のリクエスト数
		totalWeight       = 11
		socketNum         = 12
		group             sync.WaitGroup
		mutex             sync.Mutex
		ds                []Condition  // dataはオートスケールに用いる
		arrs              [13][]string // arrsはログ記録や重み調整に用いる
	)

	length := len(states)
	ors := make([]float64, length)

	for i := 0; i < 12; i++ {
		arrs[i] = make([]string, length+1)
	}
	arrs[12] = make([]string, len(sockets)+1)

	// 各配列の先頭に何の配列か記載
	arrs[operatingRatio][0] = "ors"
	arrs[cpuUsedPercent][0] = "crs"
	arrs[throughput][0] = "tps"
	arrs[dstatLog][0] = "dls"
	arrs[memoryUsedPersent][0] = "mps"
	arrs[memoryBuffer][0] = "buffers"
	arrs[memoryCached][0] = "caches"
	arrs[memoryStat][0] = "memalls"
	arrs[acquisitionTime][0] = "times"
	arrs[critical][0] = "critical"
	arrs[reqPerSec][0] = "rpss"
	arrs[totalWeight][0] = "we"
	arrs[socketNum][0] = "soc"

	arrs[totalWeight][1] = strconv.Itoa(tw)

	for i, v := range states {
		group.Add(1)
		var data Condition
		// 各サーバの付加情報毎に実行
		go func(i int, v apache.ServerStatus, sockets []apache.SocketStatus) {
			var sid apache.SocketStatus
			sid = sockets[0]
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()

			data.Name = v.HostName
			id := "[" + strconv.FormatInt(int64(v.Id), 10) + "]"

			arrs[throughput][i+1] = id + fmt.Sprintf("%f", ths[i])
			// Otherはタイムアウトした場合"Connection is timeout."が入る。
			if v.Other == "Connection is timeout." {
				// タイムアウトした場合、稼働率を1にほかを0にする。
				ors[i] = 1
				arrs[operatingRatio][i+1] = id + "1"
				for k := 1; k < 9; k++ {
					if k == 2 {
						continue
					}
					arrs[k][i+1] = id + "0"
				}
				arrs[critical][i+1] = id + v.Other
				arrs[reqPerSec][i+1] = id + "0"
				arrs[socketNum][i+1] = id + "0"
				data.Operating = 1
				data.CPU = 0
				data.Error = v.Other
				ds = append(ds, data)
			} else {
				ors[i] = v.ApacheStat
				data.Operating = ors[i]
				data.CPU = v.CpuUsedPercent[0]

				arrs[operatingRatio][i+1] = id + fmt.Sprintf("%.5f", ors[i])
				arrs[cpuUsedPercent][i+1] = id + fmt.Sprintf("%3.5f", v.CpuUsedPercent[0])
				arrs[dstatLog][i+1] = id + v.DstatLog
				arrs[memoryUsedPersent][i+1] = id + fmt.Sprintf("%3.5f", v.MemStat.UsedPercent)
				arrs[memoryBuffer][i+1] = id + fmt.Sprintf("%3.5d", v.MemStat.Buffers)
				arrs[memoryCached][i+1] = id + fmt.Sprintf("%3.5d", v.MemStat.Cached)
				arrs[memoryStat][i+1] = id + fmt.Sprint(v.MemStat)
				arrs[acquisitionTime][i+1] = id + v.Time
				arrs[reqPerSec][i+1] = id + fmt.Sprintf("%6.2f", v.ReqPerSec)
				for _, s := range sockets {
					if int64(v.Id) == int64(s.Id) {
						sid = s
						break
					}
				}
				arrs[socketNum][i+1] = id + fmt.Sprintf("%3.5d", sid.Socket)
				if ors[i] == 1 && v.CpuUsedPercent[0] >= 100 {
					arrs[critical][i+1] = id + "Operating Ratio and CPU UsedPercent is MAX!"
				}
				ds = append(ds, data)
			}
		}(i, v, sockets)
	}

	group.Wait()
	ConditionCh <- ds
	return ors, arrs
}
