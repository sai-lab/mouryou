package predictions

import (
	"time"

	"github.com/sai-lab/mouryou/lib/commands"
	"github.com/sai-lab/mouryou/lib/logger"
)

var previousTime time.Time
var afterWeight int
var oneTimeAfterLoad float64
var twoTimeAfterLoad float64

// PeriodicallyPrediction get loads after 1 hour and 2hours by TimeSeriesAnalysis.
// This methods arguments is information of the number of servers and weight.
// This methods returns the necessary weight after 1 hour.
// After 1 hour from execute TimeSeriesAnalysis, execute it once again.
// Until then this returns the previous value.
// PeriodicallyPredictionは時系列解析により1時間後と2時間後の負荷量を取得します.
// 引数はサーバ台数と重みの情報です.
// 返り値は1時間後の安定稼働に必要な重みです.
// 時系列解析を実行してから1時間経過するともう一度実行します.
// それまでは過去の値を返却します.
func PeriodicallyPrediction(w int, b int, s int, tw int, fw int) int {
	// TODO:必要な重みの計算
	var nt time.Time
	now := time.Now()

	if previousTime != nt && isNotOneTimeAfter(now) {
		return afterWeight
	}

	hls, err := commands.TimeSeriesAnalysis(now)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
		return 0
	}

	oneTimeAfterLoad = hls[0]
	twoTimeAfterLoad = hls[1]

	return 0
}

func isNotOneTimeAfter(t time.Time) bool {
	return t.Add(time.Duration(-1) * time.Hour).Before(previousTime)
}
