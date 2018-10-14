package commands

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sai-lab/mouryou/lib/logger"
)

// TimeSeriesAnalysis execute a script
// that returns loads after 1 hour and 2 hours.
func TimeSeriesAnalysis(t time.Time) ([]float64, error) {
	// TODO: edit script
	// 予測に使えるだけのデータが無かった場合のエラー処理
	hls := make([]float64, 2)
	dateparse := "1995-08-15 00:00:00"
	var start, end time.Time

	end = t.Add(time.Duration(2) * time.Hour)
	start = t.Add(time.Duration(-24) * time.Hour)

	place := logger.Place()
	logger.Debug(place, "exec TimeSeriesAnalysis script")
	out, err := exec.Command("../../modules/time_series_analysis.py",
		start.Format(dateparse), end.Format(dateparse)).Output()
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
		return nil, err
	}

	outs := strings.Split(string(out), ",")
	for i, v := range outs {
		if v == "Nil" {
			hls[i] = 0
			continue
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			place := logger.Place()
			logger.Error(place, err)
		}
		hls[i] = f
	}
	return hls, err
}
