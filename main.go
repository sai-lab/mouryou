package main

import (
	"./lib/mouryou"
	"os"
	"os/signal"
)

func main() {
	cluster := mouryou.LoadConfig()

	go mouryou.LoadMonitoringFunction(cluster)
	go mouryou.ServerManagementFunctin(cluster)
	go mouryou.DestinationSettingFunctin(cluster)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	for range sig {
		os.Exit(0)
	}
}
