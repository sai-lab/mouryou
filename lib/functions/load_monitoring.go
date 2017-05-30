package functions

import (
	"net/http"
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
		ors := OperatingRatios(sts)
		arr := convert.FloatsToStrings(ors)

		logger.Print(arr)
		logger.Write(arr)
		// logger.Send(connection, err, arr)

		loadCh <- calculate.Sum(ors)
		time.Sleep(time.Second)
	}
}

func OperatingRatios(states []apache.ServerStat) []float64 {
	ors := make([]float64, len(states))
	for i, v := range states {
		ors[i] = v.ApacheStat
	}

	return ors
}
