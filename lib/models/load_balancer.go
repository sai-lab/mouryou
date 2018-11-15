package models

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/mattn/go-pipeline"
	"github.com/sai-lab/mouryou/lib/check"
	"github.com/sai-lab/mouryou/lib/logger"
)

// LoadBalancer はロードバランサの設定情報を格納します。
type LoadBalancer struct {
	Name                            string           `json:"name"`
	VirtualIP                       string           `json:"virtual_ip"`
	Algorithm                       string           `json:"algorithm"`
	ThresholdOut                    float64          `json:"threshold_out"`
	ThresholdIn                     float64          `json:"threshold_in"`
	Margin                          float64          `json:"margin"`
	ScaleOut                        int              `json:"scale_out"`
	ScaleIn                         int              `json:"scale_in"`
	Diff                            float64          `json:"diff"`
	ThroughputAlgorithm             string           `json:"throughput_algorithm"`
	ThroughputMovingAverageInterval int64            `json:"throughput_moving_average_interval"`
	ThroughputScaleOutThreshold     int              `json:"throughput_scale_out_threshold"`
	ThroughputScaleInThreshold      int              `json:"throughput_scale_in_threshold"`
	ThroughputScaleInRatio          float64          `json:"throughput_scale_in_ratio"`
	ThroughputScaleOutRatio         float64          `json:"throughput_scale_out_ratio"`
	ThroughputScaleOutTime          int              `json:"throughput_scale_out_time"`
	ThroughputScaleInTime           int              `json:"throughput_scale_in_time"`
	UseThroughputDynamicThreshold   bool             `json:"use_throughput_dynamic_threshold"`
	ThroughputDynamicThreshold      map[string][]int `json:"throughput_dynamic_threshold"`
}

// Initialize はロードバランサの初期設定を行います。
func (lb LoadBalancer) Initialize(c *Config) {
	if c.DevelopLogLevel >= 1 {
		place := logger.Place()
		logger.Debug(place, "LoadBalancer Initialize")
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
		place := logger.Place()
		logger.Debug(place, "reload haproxy")
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
func (lb LoadBalancer) ChangeThresholdOut(working, booting, shutting, n int) {
	ocRate := float64(working+booting+shutting) / float64(n)
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

// ChangeThresholdOutInThroughputAlgorithm は起動台数に応じて閾値を切り替えます。
// 変更がない場合0.0を返します.
func (lb LoadBalancer) ChangeThresholdOutInThroughputAlgorithm(working, booting, shutting, n int) float64 {
	ocRate := (working + booting) / n * 100
	for rangeThresholdString, operatingUnitRange := range lb.ThroughputDynamicThreshold {
		if ocRate > operatingUnitRange[0] && ocRate <= operatingUnitRange[1] {
			if rangeThresholdFloat64, err := strconv.ParseFloat(rangeThresholdString, 64); err == nil {
				lb.ThroughputScaleOutRatio = rangeThresholdFloat64
				return lb.ThroughputScaleOutRatio
			} else {
				logger.Error(logger.Place(), err)
			}
		}
	}
	return 0.0
}

// ThHigh
func (balancer LoadBalancer) ThHigh(c *Config, w, n int) float64 {
	switch c.Algorithm {
	case "BasicSpike":
		return balancer.ThresholdOut
	default:
		return Threshold
	}
}

// ThLow
func (balancer LoadBalancer) ThLow(c *Config, w, n int) float64 {
	switch c.Algorithm {
	case "BasicSpike":
		return balancer.ThresholdIn * float64(w)
	default:
		return (Threshold - balancer.Diff) * (float64(w))
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
	// TODO haproxy setting
	return nil
}

// Active
func (balancer LoadBalancer) Active(name string) error {
	_, err := pipeline.Output(
		[]string{"echo", "enable", "server", "backend_servers/" + name},
		[]string{"socat", "stdio", "/tmp/haproxy-cli.sock"},
	)
	if err != nil {
		return err
	}
	return nil
}

// Inactive
func (balancer LoadBalancer) Inactive(name string) error {
	_, err := pipeline.Output(
		[]string{"echo", "disable", "server", "backend_servers/" + name},
		[]string{"socat", "stdio", "/tmp/haproxy-cli.sock"},
	)
	if err != nil {
		return err
	}
	return nil
}

// ChangeWeight
func (balancer LoadBalancer) ChangeWeight(name string, weight int) error {
	_, err := pipeline.Output(
		[]string{"echo", "set", "weight", "backend_servers/" + name, strconv.FormatInt(int64(weight), 10)},
		[]string{"socat", "stdio", "/tmp/haproxy-cli.sock"},
	)
	if err != nil {
		return err
	}
	return nil
}
