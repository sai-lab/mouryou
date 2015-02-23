package main

import (
	"./lib/tenbin"
	"os"
	"os/signal"
	"time"
)

func main() {
	hypervisor := tenbin.LoadConfig()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for range sig {
			os.Exit(0)
		}
	}()

	for {
		hypervisor.PrintLoads()
		time.Sleep(time.Second)
	}
}
