package tenbin

import (
	"../apache"
)

type virtualMachine struct {
	Name string
	Host string
}

func (vm virtualMachine) operatingRatio() float64 {
	board, err := apache.Scoreboard(vm.Host)

	if err != nil {
		switch err.Error() {
		case "apache: no response":
			return 0.0
		case "apache: request timeout":
			return 1.0
		}
	}

	return apache.OperatingRatio(board)
}
