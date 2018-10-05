package monitor

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

func LoadMonitoring(config *models.Config) {
	http.DefaultClient.Timeout = time.Duration(config.Timeout * time.Second)
	connection, err := config.WebSocket.Dial()

	for {
		bootedServers := []string{}
		//
		for _, v := range GetStates() {
			if config.DevelopLogLevel >= 6 {
				logger.PrintPlace("GetStates() Machine Name: " + v.Name + "Machine Info: " + v.Info)
			}
			if v.Info == "booted up" {
				bootedServers = append(bootedServers, v.Name)
			}
		}

		satuses := config.Cluster.ServerStatuses(bootedServers)
		ors, arrs := Ratios(satuses)

		logger.PWArrays(config.DevelopLogLevel, arrs)
		logger.Send(connection, err, arrs)

		LoadCh <- calculate.Sum(ors)
		time.Sleep(time.Second)
	}
}

func Ratios(states []apache.ServerStatus) ([]float64, [11][]string) {
	var (
		operatingRatio    = 0
		cpuUsedPercent    = 1
		throughput        = 2 // throughputs is ApacheLog in apache.ServerStatus
		dstatLog          = 3
		memoryUsedPersent = 4
		memoryBuffer      = 5
		memoryCached      = 6
		memoryStat        = 7
		acquisitionTime   = 8
		critical          = 9
		reqPerSec         = 10 // reqPerSecは起動してからの平均の1秒間のリクエスト数
		group             sync.WaitGroup
		mutex             sync.Mutex
		ds                []Data       // dataはオートスケールに用いる
		arrs              [11][]string // arrsはログ記録や重み調整に用いる
	)

	length := len(states)
	ors := make([]float64, length)

	for i := 0; i < 11; i++ {
		arrs[i] = make([]string, length+1)
	}

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

	for i, v := range states {
		group.Add(1)
		var data Data
		// 各サーバの付加情報毎に実行
		go func(i int, v apache.ServerStatus) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()

			data.Name = v.HostName
			id := "[" + strconv.FormatInt(int64(v.Id), 10) + "]"

			// Otherはタイムアウトした場合"Connection is timeout."が入る。
			if v.Other == "Connection is timeout." {
				// タイムアウトした場合、稼働率を1にほかを0にする。
				ors[i] = 1
				arrs[operatingRatio][i+1] = id + "1"
				for k := 1; k < 9; k++ {
					arrs[k][i+1] = id + "0"
				}
				arrs[critical][i+1] = id + v.Other
				arrs[reqPerSec][i+1] = id + "0"
				data.Operating = 1
				data.CPU = 0
				data.Throughput = 0
				data.Error = v.Other
				ds = append(ds, data)
			} else {
				if beforeTime[v.HostName] == 0 {
					data.Throughput = 0
				} else {
					data.Throughput = (v.ApacheLog - beforeTotalAccess[v.HostName]) / (int(time.Now().Unix()) - beforeTime[v.HostName])
				}

				beforeTime[v.HostName] = int(time.Now().Unix())
				beforeTotalAccess[v.HostName] = v.ApacheLog
				beforeTime[v.HostName] = int(time.Now().Unix())

				ors[i] = v.ApacheStat
				data.Operating = ors[i]
				data.CPU = v.CpuUsedPercent[0]

				arrs[operatingRatio][i+1] = id + fmt.Sprintf("%.5f", ors[i])
				arrs[cpuUsedPercent][i+1] = id + fmt.Sprintf("%3.5f", v.CpuUsedPercent[0])
				arrs[throughput][i+1] = id + fmt.Sprintf("%5d", v.ApacheLog)
				arrs[dstatLog][i+1] = id + v.DstatLog
				arrs[memoryUsedPersent][i+1] = id + fmt.Sprintf("%3.5f", v.MemStat.UsedPercent)
				arrs[memoryBuffer][i+1] = id + fmt.Sprintf("%3.5d", v.MemStat.Buffers)
				arrs[memoryCached][i+1] = id + fmt.Sprintf("%3.5d", v.MemStat.Cached)
				arrs[memoryStat][i+1] = id + fmt.Sprint(v.MemStat)
				arrs[acquisitionTime][i+1] = id + v.Time
				arrs[reqPerSec][i+1] = id + fmt.Sprintf("%6.2f", v.ReqPerSec)
				if ors[i] == 1 && v.CpuUsedPercent[0] >= 100 {
					arrs[critical][i+1] = id + "Operating Ratio and CPU UsedPercent is MAX!"
				}
				ds = append(ds, data)
			}
		}(i, v)
	}

	group.Wait()
	DataCh <- ds
	return ors, arrs
}
