package main

import (
	"./lib/apache"
	"fmt"
)

func main() {
	board := apache.Scoreboard("192.168.11.21")
	fmt.Println(apache.OperatingRatio(board))
}
