package main

import (
	"./lib/mouryou"
	"log"
	"os"
	"os/signal"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	cluster := mouryou.LoadConfig()
	f := mouryou.CreateLog()
	log.SetOutput(f)

	go mouryou.LoadMonitoringFunction(cluster)
	go mouryou.ServerManagementFunctin(cluster)
	go mouryou.DestinationSettingFunctin(cluster)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for range sig {
		f.Close()
		os.Exit(0)
	}
}
