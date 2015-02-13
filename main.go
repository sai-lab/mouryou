package main

import (
	"./apache"
	"fmt"
)

func main() {
	board := apache.GetScoreboard("192.168.11.21", 80)
	fmt.Println(apache.OperatingRatio(board))
}
