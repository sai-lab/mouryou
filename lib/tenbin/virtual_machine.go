package tenbin

import (
	"../apache"
)

type virtualMachine struct {
	name      string
	ipAddress string
}

func (vm virtualMachine) operatingRatio() float64 {
	board := apache.Scoreboard(vm.ipAddress)
	return apache.OperatingRatio(board)
}
