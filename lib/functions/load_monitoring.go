package functions

import (
	"net/http"
	"time"

	"golang.org/x/net/websocket"

	"github.com/sai-lab/mouryou/lib/average"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func LoadMonitoring(config *models.ConfigStruct, ws *websocket.Conn) {
	var w int

	http.DefaultClient.Timeout = time.Duration(config.Timeout * time.Second)

	for {
		w = mutex.Read(&working, &workMutex)
		ors := config.Cluster.OperatingRatios(w)
		arr := convert.FloatsToStrings(ors)

		logger.Print(arr)
		logger.Write(arr)
		if ws != nil {
			logger.Send(ws, arr)
		}

		loadCh <- average.Average(ors)
		time.Sleep(time.Second)
	}
}
