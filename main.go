package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"

	"./lib/mouryou"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cluster := mouryou.LoadConfig()

	f := mouryou.CreateLog()
	log.SetOutput(f)
	log.SetFlags(log.Ltime)

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
