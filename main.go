package main

import (
	"./lib/mouryou"
	"os"
	"os/signal"
	"time"
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

	for {
		cluster.Log()
		time.Sleep(time.Second)
	}
}
