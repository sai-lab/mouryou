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

func LoadMonitoring(config *models.ConfigStruct) {
	var mu sync.RWMutex

	http.DefaultClient.Timeout = time.Duration(config.Timeout * time.Second)
	// connection, err := config.WebSocket.Dial()

	for {
		mu.RLock()
		status := States
		mu.RUnlock()

		bt := []string{}
		for _, v := range status {
			if v.Info == "booted up" {
				bt = append(bt, v.Name)
			}
		}

		sts := config.Cluster.ServerStates(bt)
		ors, arrs := Ratios(sts)

		logger.PWArrays(arrs)
		// logger.Send(connection, err, arr)

		LoadCh <- calculate.Sum(ors)
		time.Sleep(time.Second)
	}
}

func Ratios(states []apache.ServerStat) ([]float64, [11][]string) {
	var group sync.WaitGroup
	var mutex sync.Mutex
	var ds []DataStruct

	length := len(states)
	ors := make([]float64, length)
	var arrs [11][]string

	for i := 0; i < 11; i++ {
		arrs[i] = make([]string, length+1)
	}

	arrs[0][0] = "ors"
	arrs[1][0] = "crs"
	arrs[2][0] = "tps"
	arrs[3][0] = "dls"
	arrs[4][0] = "mps"
	arrs[5][0] = "buffers"
	arrs[6][0] = "caches"
	arrs[7][0] = "memalls"
	arrs[8][0] = "times"
	arrs[9][0] = "critical"
	arrs[10][0] = "rpss"

	for i, v := range states {
		group.Add(1)
		var data DataStruct
		go func(i int, v apache.ServerStat) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()

			data.Name = v.HostName
			id := "[" + strconv.FormatInt(int64(v.Id), 10) + "]"

			if v.Other != "" {
				ors[i] = 1
				arrs[0][i+1] = id + "1"
				for k := 1; k < 9; k++ {
					arrs[k][i+1] = id + "0"
				}
				arrs[9][i+1] = id + v.Other
				arrs[10][i+1] = id + "0"
				data.Operating = 1
				data.Cpu = 0
				data.Throughput = 0
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
				data.Cpu = v.CpuUsedPercent[0]

				ds = append(ds, data)
				arrs[0][i+1] = id + fmt.Sprintf("%.5f", ors[i])
				arrs[1][i+1] = id + fmt.Sprintf("%3.5f", v.CpuUsedPercent[0])
				arrs[2][i+1] = id + fmt.Sprintf("%5d", v.ApacheLog)
				arrs[3][i+1] = id + v.DstatLog
				arrs[4][i+1] = id + fmt.Sprintf("%3.5f", v.MemStat.UsedPercent)
				arrs[5][i+1] = id + fmt.Sprintf("%3.5d", v.MemStat.Buffers)
				arrs[6][i+1] = id + fmt.Sprintf("%3.5d", v.MemStat.Cached)
				arrs[7][i+1] = id + fmt.Sprint(v.MemStat)
				arrs[8][i+1] = id + v.Time
				arrs[10][i+1] = id + fmt.Sprintf("%6.2f", v.ReqPerSec)
				if ors[i] == 1 && v.CpuUsedPercent[0] >= 100 {
					arrs[9][i+1] = "[" + id + "]" + "Operating Ratio and CPU UsedPercent is MAX!"
				}
			}
		}(i, v)
	}
	group.Wait()
	DataCh <- ds
	return ors, arrs
}
