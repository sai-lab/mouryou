package functions

import (
	"net/http"
	"sync"
	"time"

	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func LoadMonitoring(config *models.ConfigStruct) {
	var w int

	http.DefaultClient.Timeout = time.Duration(config.Timeout * time.Second)
	// connection, err := config.WebSocket.Dial()

	for {
		w = mutex.Read(&working, &workMutex)
		sts := config.Cluster.ServerStates(w)
		ors, crs := Ratios(sts)
		orArr := convert.FloatsToStrings(ors, "ors")
		crArr := convert.FloatsToStrings(crs, "crs")

		logger.Print(orArr)
		logger.Write(orArr)
		logger.Print(crArr)
		logger.Write(crArr)
		// logger.Send(connection, err, arr)

		loadCh <- calculate.Sum(ors)
		time.Sleep(time.Second)
	}
}

func Ratios(states []apache.ServerStat) ([]float64, []float64) {
	var group sync.WaitGroup
	var mutex sync.Mutex

	ors := make([]float64, len(states))
	crs := make([]float64, len(states))

	for i, v := range states {
		group.Add(1)
		go func(i int, v apache.ServerStat) {
			defer group.Done()
			mutex.Lock()
			defer mutex.Unlock()

			if v.Other != "" {
				logger.PrintPlace(v.HostName + " Other error is occured! : " + v.Other)
				ors[i] = 1
				crs[i] = 0
			} else {
				ors[i] = v.ApacheStat
				crs[i] = v.CpuUsedPercent[0]
				if ors[i] == 1 && crs[i] == 100 {
					models.CriticalCh <- v.HostName
				}
			}
		}(i, v)
	}
	group.Wait()

	return ors, crs
}
