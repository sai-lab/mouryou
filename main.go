package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/sai-lab/mouryou/lib/databases"
	"github.com/sai-lab/mouryou/lib/engine"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初期設定
	c := new(models.Config)
	c.LoadSetting(os.Getenv("HOME") + "/.mouryou.json")
	startServers := c.Cluster.Initialize(c)
	engine.ServerWeightInitialize(c, len(startServers))

	// InfluxDBに接続
	err := databases.Connect(c)
	if err != nil {
		panic(err)
	}

	// データベースを作成
	_, err = databases.QueryDB(c.InfluxDBConnection, fmt.Sprintf("CREATE DATABASE %s", c.InfluxDBServerDB), "")
	if err != nil {
		panic(err)
	}

	// ログ出力設定
	file := logger.Create()
	log.SetOutput(file)
	log.SetFlags(log.Ltime)

	go monitor.LoadMonitoring(c)
	go engine.LoadDetermination(c)
	go engine.ServerManagement(c)
	go engine.DestinationSetting(c)
	go engine.ServerStatesManager()
	//go engine.WeightOperator(c)
	go engine.WeightManager(c)
	go monitor.WeightMonitor(c)
	go monitor.MeasureServer(c)

	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)

	for range channel {
		file.Close()
		os.Exit(0)
	}
}
