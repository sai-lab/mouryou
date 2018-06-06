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

	c := new(models.Config)
	c.LoadSetting(os.Getenv("HOME") + "/.mouryou.json")
	c.Cluster.Initialize(c)
	engine.Initialize(c)

	file := logger.Create()
	log.SetOutput(file)
	log.SetFlags(log.Ltime)

	go monitor.LoadMonitoring(c)
	go engine.ServerManagement(c)
	go engine.DestinationSetting(c)
	go engine.StatusManager()
	//go engine.WeightOperator(c)
	if c.Algorithm == "DecreaseWeightFromBasicSpike" {
		go engine.DecreaseWeight(c)
	}
	go monitor.MeasureServer(c)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	for range channel {
		file.Close()
		os.Exit(0)
	}
}
