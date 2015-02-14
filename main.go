package main

import (
	"./lib/apache"
	"./lib/tenbin"
	"fmt"
	"strconv"
)

func main() {
	var hypervisor tenbin.Hypervisor
	apache.Timeout = 1

	for i := 1; i < 10; i++ {
		num := strconv.Itoa(i)
		hypervisor.AddVM("web-server-"+num, "192.168.11.2"+num)
	}

	fmt.Printf("%+v\n", hypervisor)
	fmt.Printf("%+v\n", hypervisor.AVGOR())
}
