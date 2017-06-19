package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/sai-lab/mouryou/lib/functions"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := models.LoadConfig(os.Getenv("HOME") + "/.mouryou.json")
	config.Cluster.Initialize()
	functions.Initialize(config)

	file := logger.Create()
	log.SetOutput(file)
	log.SetFlags(log.Ltime)

	go functions.LoadMonitoring(config)
	go functions.ServerManagement(config)
	go functions.DestinationSetting(config)
	go functions.StatusManager()
	// go functions.MonitorWeightChange(config)
	go functions.MeasureServer(config)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	for range channel {
		file.Close()
		os.Exit(0)
	}
}
