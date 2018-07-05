package models

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	pipeline "github.com/mattn/go-pipeline"
	"github.com/sai-lab/mouryou/lib/check"
	"github.com/sai-lab/mouryou/lib/logger"
)

// LoadBalancer はロードバランサの設定情報を格納します。
type LoadBalancer struct {
	Name         string  `json:"name"`
	VirtualIP    string  `json:"virtual_ip"`
	Algorithm    string  `json:"algorithm"`
	ThresholdOut float64 `json:"threshold_out"`
	ThresholdIn  float64 `json:"threshold_in"`
	Margin       float64 `json:"margin"`
	ScaleOut     int     `json:"scale_out"`
	ScaleIn      int     `json:"scale_in"`
	Diff         float64 `json:"diff"`
}

// Initialize はロードバランサの初期設定を行います。
func (lb LoadBalancer) Initialize(c *Config) {
	if c.DevelopLogLevel >= 1 {
		logger.PrintPlace("Load Balancer Initialize")
	}
	// 仮想IPを設定します。
	exec.Command("ip", "addr", "add", lb.VirtualIP, "label", "eth0:vip", "dev", "eth0").Run()

	switch lb.Name {
	case "ipvs":
		err := exec.Command("ipvsadm", "-C").Run()
		check.Error(err)
		err = exec.Command("ipvsadm", "-A", "-t", lb.VirtualIP+":http", "-s", lb.Algorithm).Run()
		check.Error(err)
	case "haproxy":
		logger.PrintPlace("reload haproxy")
		err := exec.Command("systemctl", "reload", "haproxy").Run()
		check.Error(err)
	default:
		err := errors.New("cannot determine the name of load balancer")
		check.Error(err)
	}

	time.Sleep(time.Duration(1) * 3)
}

func (lb *LoadBalancer) valueCheck() error {
	var err error
	var errTxt string
	var e string

	if lb.Name == "" {
		e = "please set load_balancer name for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if lb.VirtualIP == "" {
		e = "please set load_balancer virtual_ip for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if lb.Algorithm == "" {
		e = "please set load_balancer algorithm for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if lb.ThresholdOut == float64(0) {
		e = "please set load_balancer threshold_out for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if lb.ThresholdIn == float64(0) {
		e = "please set load_balancer timeout value for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}

	if errTxt != "" {
		err = errors.New(errTxt)
	}

	return err
}

// ChangeThresholdOut は起動台数に応じて閾値を切り替えます。
func (lb LoadBalancer) ChangeThresholdOut(working, booting, shuting, n int) {
	var ocRate float64
	ocRate = float64(working+booting+shuting) / float64(n)
	ocRateLog := []string{"ocRateLog", fmt.Sprintf("%5.3f %d %d %d %d", ocRate, working, booting, shuting, n)}
	logger.Write(ocRateLog)
	switch {
	case ocRate <= 0.3:
		Threshold = 0.1
	case ocRate <= 0.5:
		Threshold = 0.3
	case ocRate <= 0.7:
		Threshold = 0.5
	case ocRate <= 0.9:
		Threshold = 0.6
	case ocRate <= 1.0:
		Threshold = 0.7
	}
}

// ThHigh
func (balancer LoadBalancer) ThHigh(c *Config, w, n int) float64 {
	switch c.Algorithm {
	case "basic_spike":
		return balancer.ThresholdOut
	default:
		return Threshold
	}
}

// ThLow
func (balancer LoadBalancer) ThLow(c *Config, w, n int) float64 {
	switch c.Algorithm {
	case "basic_spike":
		return balancer.ThresholdOut*float64(w) - balancer.Margin
	default:
		//return (Threshold-balancer.Diff)*(float64(w)) - balancer.Margin
		return (Threshold-balancer.Diff)*float64(w/n) - balancer.Margin
	}
}

// Add
func (balancer LoadBalancer) Add(name string) error {
	// err := exec.Command("ipvsadm", "-a", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "0", "-g").Run()
	// if err != nil {
	// 	return err
	// }

	return nil
}

// Remove
func (balancer LoadBalancer) Remove(name string) error {
	// err := exec.Command("ipvsadm", "-d", "-t", balancer.VirtualIP+":http", "-r", host+":http").Run()
	// if err != nil {
	// 	return err
	// }

	// TODO haproxy setting

	return nil
}

// Active
func (balancer LoadBalancer) Active(name string) error {
	// err := exec.Command("ipvsadm", "-e", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "1", "-g").Run()

	logger.PrintPlace("enable server " + fmt.Sprint(name))
	logger.WriteMonoString("enable server " + fmt.Sprint(name))
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

// Inactive
func (balancer LoadBalancer) Inactive(name string) error {
	//err := exec.Command("ipvsadm", "-e", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "0", "-g").Run()

	logger.PrintPlace("disable server " + fmt.Sprint(name))
	logger.WriteMonoString("disable server " + fmt.Sprint(name))
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

// ChangeWeight
func (balancer LoadBalancer) ChangeWeight(name string, weight int) error {
	logger.PrintPlace("change server weight " + fmt.Sprint(name) + ", " + fmt.Sprint(weight))
	logger.WriteMonoString("change server weight " + fmt.Sprint(name) + ", " + fmt.Sprint(weight))
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
