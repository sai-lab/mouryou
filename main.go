package main

import (
	"./lib/tenbin"
	"os"
	"os/signal"
	"time"
)

func main() {
	cluster := tenbin.LoadConfig()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for range sig {
			os.Exit(0)
		}
	}()

	for {
		cluster.Log()
		time.Sleep(time.Second)
	}
}
