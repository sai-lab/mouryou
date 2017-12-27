package predictions

import (
	"fmt"
	"time"

	"github.com/sai-lab/mouryou/lib/commands"
	"github.com/sai-lab/mouryou/lib/logger"
)

var previousTime time.Time
var afterWeight int
var oneTimeAfterLoad float64
var twoTimeAfterLoad float64

// PeriodicallyPrediction get loads after 1 hour and 2hours by TimeSeriesAnalysis.
// This arguments is information of the number of servers and weight.
// This returns the necessary weight after 1 hour.
// After 1 hour from execute TimeSeriesAnalysis, execute it once again.
// Until then this returns the previous value
func PeriodicallyPrediction(w int, b int, s int, tw int) int {
	// TODO:必要な重みの計算
	var nt time.Time
	now := time.Now()

	if previousTime != nt && isNotOneTimeAfter(now) {
		return afterWeight
	}

	hls, err := commands.TimeSeriesAnalysis(now)
	if err != nil {
		logger.PrintPlace(fmt.Sprint(err))
		return 0
	}

	oneTimeAfterLoad = hls[0]
	twoTimeAfterLoad = hls[1]

	return 0
}

func isNotOneTimeAfter(t time.Time) bool {
	return t.Add(time.Duration(-1) * time.Hour).Before(previousTime)
}
