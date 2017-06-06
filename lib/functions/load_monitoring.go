package functions

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
		status := states
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

		loadCh <- calculate.Sum(ors)
		time.Sleep(time.Second)
	}
}

func Ratios(states []apache.ServerStat) ([]float64, [7][]string) {
	var group sync.WaitGroup
	var mutex sync.Mutex

	length := len(states)
	ors := make([]float64, length)
	var arrs [7][]string

	for i := 0; i < 7; i++ {
		arrs[i] = make([]string, length+1)
	}

	arrs[0][0] = "ors"
	arrs[1][0] = "crs"
	arrs[2][0] = "tps"
	arrs[3][0] = "dls"
	arrs[4][0] = "mps"
	arrs[5][0] = "times"
	arrs[6][0] = "critical"

	for i, v := range states {
		group.Add(1)
		go func(i int, v apache.ServerStat) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()

			id := strconv.FormatInt(int64(v.Id), 10)
			if v.Other != "" {
				logger.PrintPlace(v.HostName + " Other error is occured! : " + v.Other)
				ors[i] = 1
				arrs[0][i+1] = "[" + id + "]" + "1"
				arrs[1][i+1] = "[" + id + "]" + "0"
				arrs[2][i+1] = "[" + id + "]" + "0"
				arrs[3][i+1] = "[" + id + "]" + "0"
				arrs[4][i+1] = "[" + id + "]" + "0"
				arrs[5][i+1] = "[" + id + "]" + "0"
			} else {
				ors[i] = v.ApacheStat
				arrs[0][i+1] = "[" + id + "]" + fmt.Sprintf("%.5f", ors[i])
				arrs[1][i+1] = "[" + id + "]" + fmt.Sprintf("%3.5f", v.CpuUsedPercent[0])
				arrs[2][i+1] = "[" + id + "]" + fmt.Sprintf("%5d", v.ApacheLog)
				arrs[3][i+1] = "[" + id + "]" + v.DstatLog
				arrs[4][i+1] = "[" + id + "]" + fmt.Sprintf("%3.5f", v.UsedPercent)
				arrs[5][i+1] = "[" + id + "]" + v.Time
				if ors[i] == 1 && v.CpuUsedPercent[0] >= 100 {
					fmt.Println("critical is occured in " + v.HostName)
					if criticalCh != nil {
						arrs[6][i+1] = "[" + id + "]" + "Operating Ratio and CPU UsedPercent is MAX!"
						c := CriticalStruct{v.HostName, "critical"}
						criticalCh <- c
					}
				}
			}
		}(i, v)
	}
	group.Wait()

	return ors, arrs
}
