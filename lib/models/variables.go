package models

var (
	threshold float64
	vmQue     = make([]chan bool, 11)
)
