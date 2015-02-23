package main

import (
	"./lib/apache"
	"./lib/tenbin"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	var hypervisor tenbin.Hypervisor
	apache.Timeout = 1
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		for range sig {
			os.Exit(0)
		}
	}()

	for i := 1; i < 10; i++ {
		num := strconv.Itoa(i)
		hypervisor.AddVM("web-server-"+num, "192.168.11.2"+num)
	}

	for {
		hypervisor.PrintLoads()
		time.Sleep(time.Second)
	}
}
