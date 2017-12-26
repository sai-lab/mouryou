package models

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	pipeline "github.com/mattn/go-pipeline"
	"github.com/sai-lab/mouryou/lib/check"
	"github.com/sai-lab/mouryou/lib/logger"
)

type LoadBalancerStruct struct {
	VirtualIP    string  `json:"virtual_ip"`
	Algorithm    string  `json:"algorithm"`
	ThresholdOut float64 `json:"threshold_out"`
	ThresholdIn  float64 `json:"threshold_in"`
	Margin       float64 `json:"margin"`
	ScaleOut     int     `json:"scale_out"`
	ScaleIn      int     `json:"scale_in"`
	Diff         float64 `json:"diff"`
}

func (balancer LoadBalancerStruct) Initialize() {
	logger.PrintPlace("Load Balancer Initialize")
	exec.Command("ip", "addr", "add", balancer.VirtualIP, "label", "eth0:vip", "dev", "eth0").Run()

	// err := exec.Command("ipvsadm", "-C").Run()
	// check.Error(err)
	// err = exec.Command("ipvsadm", "-A", "-t", balancer.VirtualIP+":http", "-s", balancer.Algorithm).Run()
	// check.Error(err)

	logger.PrintPlace("reload haproxy")
	err := exec.Command("systemctl", "reload", "haproxy").Run()
	check.Error(err)

	time.Sleep(time.Duration(1) * 3)
}

// func (balancer LoadBalancerStruct) ChangeThresholdOut(w, b, s, n int) {
// 	var ocRate float64
// 	ocRate = float64(w+b+s) / float64(n)
// 	switch {
// 	case ocRate <= 0.3:
// 		Threshold = 0.5
// 	case ocRate <= 0.5:
// 		Threshold = 0.6
// 	case ocRate <= 0.7:
// 		Threshold = 0.7
// 	case ocRate <= 0.9:
// 		Threshold = 0.8
// 	case ocRate <= 1.0:
// 		Threshold = 0.9
// 	}
// }

func (balancer LoadBalancerStruct) ThHigh(w, n int) float64 {
	return Threshold
	// return balancer.ThresholdOut
}

func (balancer LoadBalancerStruct) ThLow(w int) float64 {
	return (Threshold-balancer.Diff)*float64(w) - balancer.Margin
	// return balancer.ThresholdIn*float64(w) - balancer.Margin
}

func (balancer LoadBalancerStruct) Add(name string) error {
	// err := exec.Command("ipvsadm", "-a", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "0", "-g").Run()
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (balancer LoadBalancerStruct) Remove(name string) error {
	// err := exec.Command("ipvsadm", "-d", "-t", balancer.VirtualIP+":http", "-r", host+":http").Run()
	// if err != nil {
	// 	return err
	// }

	// TODO haproxy setting

	return nil
}

func (balancer LoadBalancerStruct) Active(name string) error {
	// err := exec.Command("ipvsadm", "-e", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "1", "-g").Run()

	logger.PrintPlace("enable server " + fmt.Sprint(name))
	_, err := pipeline.Output(
		[]string{"echo", "enable", "server", "backend_servers/" + name},
		[]string{"socat", "stdio", "/tmp/haproxy-cli.sock"},
	)

	if err != nil {
		logger.PrintPlace(fmt.Sprint(err))
		return err
	}

	return nil
}

func (balancer LoadBalancerStruct) Inactive(name string) error {
	//err := exec.Command("ipvsadm", "-e", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "0", "-g").Run()

	logger.PrintPlace("disable server " + fmt.Sprint(name))
	_, err := pipeline.Output(
		[]string{"echo", "disable", "server", "backend_servers/" + name},
		[]string{"socat", "stdio", "/tmp/haproxy-cli.sock"},
	)
	if err != nil {
		logger.PrintPlace(fmt.Sprint(err))
		return err
	}

	return nil
}

func (balancer LoadBalancerStruct) ChangeWeight(name string, weight int) error {
	fmt.Println("hoge")
	logger.PrintPlace("change server weight " + fmt.Sprint(name) + ", " + fmt.Sprint(weight))
	_, err := pipeline.Output(
		[]string{"echo", "set", "weight", "backend_servers/" + name, strconv.FormatInt(int64(weight), 10)},
		[]string{"socat", "stdio", "/tmp/haproxy-cli.sock"},
	)
	if err != nil {
		logger.PrintPlace(fmt.Sprint(err))
		return err
	}

	return nil

}
