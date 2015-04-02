package main

import (
	"./lib/mouryou"
	"os"
	"os/signal"
)

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for range sig {
			os.Exit(0)
		}
	}()

	cluster := mouryou.LoadConfig()
	cluster.Run()
}
