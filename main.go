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
	cluster.Initialize()

	file := mouryou.CreateLog()
	log.SetOutput(file)
	log.SetFlags(log.Ltime)

	go mouryou.LoadMonitoringFunction(cluster)
	go mouryou.ServerManagementFunctin(cluster)
	go mouryou.DestinationSettingFunctin(cluster)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	for range channel {
		file.Close()
		os.Exit(0)
	}
}
