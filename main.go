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

	cluster := models.LoadConfig(os.Getenv("HOME") + "/.mouryou.json")
	cluster.Initialize()

	file := logger.Create()
	log.SetOutput(file)
	log.SetFlags(log.Ltime)

	go functions.LoadMonitoring(cluster)
	go functions.ServerManagement(cluster)
	go functions.DestinationSetting(cluster)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	for range channel {
		file.Close()
		os.Exit(0)
	}
}
