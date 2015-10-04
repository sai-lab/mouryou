package functions

import (
	"net/http"
	"time"

	"github.com/sai-lab/mouryou/lib/average"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func LoadMonitoring(cluster *models.ClusterStruct) {
	var w int

	http.DefaultClient.Timeout = time.Duration(TIMEOUT_SEC * time.Second)

	for {
		w = mutex.Read(&working, &workMutex)
		ors := cluster.OperatingRatios(w)

		logger.Print(ors)
		logger.Write(ors)

		loadCh <- average.Average(ors)
		time.Sleep(time.Second)
	}
}
