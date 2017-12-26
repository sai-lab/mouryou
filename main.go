package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/sai-lab/mouryou/lib/engine"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := models.LoadConfig(os.Getenv("HOME") + "/.mouryou.json")
	config.Cluster.Initialize()
	engine.Initialize(config)

	file := logger.Create()
	log.SetOutput(file)
	log.SetFlags(log.Ltime)

	go monitor.LoadMonitoring(config)
	go engine.ServerManagement(config)
	go engine.DestinationSetting(config)
	go engine.StatusManager()
	go engine.WeightOperator(config)
	go monitor.MeasureServer(config)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	for range channel {
		file.Close()
		os.Exit(0)
	}
}
