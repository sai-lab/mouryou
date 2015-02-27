package main

import (
	"./lib/tenbin"
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

	cluster := tenbin.LoadConfig()
	cluster.InitVMs()

	for {
		cluster.Log()
		time.Sleep(time.Second)
	}
}
