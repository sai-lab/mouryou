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
	Name                               string           `json:"name"`
	VirtualIP                          string           `json:"virtual_ip"`
	LoadBalancingAlgorithm             string           `json:"load_balancing_algorithm"`
	OperatingRatioAlgorithm            string           `json:"operating_ratio_algorithm"`
	OperatingRatioThresholdOut         float64          `json:"operating_ratio_threshold_out"`
	OperatingRatioThresholdIn          float64          `json:"operating_ratio_threshold_in"`
	OperatingRatioMargin               float64          `json:"operating_ratio_margin"`
	OperatingRatioScaleOutInterval     int              `json:"operating_ratio_scale_out_interval"`
	OperatingRatioScaleInInterval      int              `json:"operating_ratio_scale_in_interval"`
	OperatingRatioDynamicThresholdDiff float64          `json:"operating_ratio_dynamic_threshold_diff"`
	OperatingRatioDynamicThreshold     map[string][]int `json:"operating_ratio_dynamic_threshold"`
	ThroughputAlgorithm                string           `json:"throughput_algorithm"`
	ThroughputMovingAverageInterval    int64            `json:"throughput_moving_average_interval"`
	ThroughputScaleOutThreshold        int              `json:"throughput_scale_out_threshold"`
	ThroughputScaleInThreshold         int              `json:"throughput_scale_in_threshold"`
	ThroughputScaleInRatio             float64          `json:"throughput_scale_in_ratio"`
	ThroughputScaleOutRatio            float64          `json:"throughput_scale_out_ratio"`
	ThroughputScaleOutTime             int              `json:"throughput_scale_out_time"`
	ThroughputScaleInTime              int              `json:"throughput_scale_in_time"`
	UseThroughputDynamicThreshold      bool             `json:"use_throughput_dynamic_threshold"`
	ThroughputScaleOutRate             float64          `json:"throughput_scale_out_rate"`
	ThroughputScaleInRate              float64          `json:"throughput_scale_in_rate"`
	ThroughputDynamicThreshold         map[string][]int `json:"throughput_dynamic_threshold"`
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
		err = exec.Command("ipvsadm", "-A", "-t", lb.VirtualIP+":http", "-s", lb.LoadBalancingAlgorithm).Run()
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

	time.Sleep(time.Duration(3) * time.Second)
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
	if lb.LoadBalancingAlgorithm == "" {
		e = "please set load_balancer algorithm for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if lb.OperatingRatioThresholdOut == float64(0) {
		e = "please set load_balancer threshold_out for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if lb.OperatingRatioThresholdIn == float64(0) {
		e = "please set load_balancer timeout value for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}

	if errTxt != "" {
		err = errors.New(errTxt)
	}

	return err
}

// ChangeThresholdOutInOperatingRatioAlgorithm は起動台数に応じて閾値を切り替えます。
func (lb LoadBalancer) ChangeThresholdOutInOperatingRatioAlgorithm(working, booting, n int) (float64, int) {
	ocRate := int(float32(working+booting) / float32(n) * 100.0)
	for rangeThresholdString, operatingUnitRange := range lb.OperatingRatioDynamicThreshold {
		if ocRate > operatingUnitRange[0] && ocRate <= operatingUnitRange[1] {
			if rangeThresholdFloat64, err := strconv.ParseFloat(rangeThresholdString, 64); err == nil {
				Threshold = rangeThresholdFloat64
				return Threshold, ocRate
			} else {
				logger.Error(logger.Place(), err)
			}
		}
	}
	return 0.0, ocRate
}

// ChangeThresholdOutInThroughputAlgorithm は起動台数に応じて閾値を切り替えます。
// 変更がない場合0.0を返します.
func (lb LoadBalancer) ChangeThresholdOutInThroughputAlgorithm(working, booting, n int) (float64, int) {
	ocRate := int(float32(working+booting) / float32(n) * 100.0)
	for rangeThresholdString, operatingUnitRange := range lb.ThroughputDynamicThreshold {
		if ocRate > operatingUnitRange[0] && ocRate <= operatingUnitRange[1] {
			if rangeThresholdFloat64, err := strconv.ParseFloat(rangeThresholdString, 64); err == nil {
				lb.ThroughputScaleOutRatio = rangeThresholdFloat64
				return lb.ThroughputScaleOutRatio, ocRate
			} else {
				logger.Error(logger.Place(), err)
			}
		}
	}
	return 0.0, ocRate
}

// ChangeThresholdOutInThroughputAlgorithm は起動台数に応じて閾値を切り替えます。
// 変更がない場合0.0を返します.
func (lb LoadBalancer) ChangeThresholdOutInThroughput(working, booting, n int) (float64, float64) {
	lb.ThroughputScaleOutRatio = float64((float64(working+booting) - lb.ThroughputScaleOutRate) / float64(working+booting))
	lb.ThroughputScaleInRatio = float64(float64(float64(working+booting)-1-lb.ThroughputScaleInRate) / float64(working+booting))
	if lb.ThroughputScaleInRatio < 0.0 {
		lb.ThroughputScaleInRatio = 0.0
	}
	return lb.ThroughputScaleOutRatio, lb.ThroughputScaleInRatio
}

// ThHighInOperatingRatioAlgorithm は稼働率ベースのアルゴリズムで使われる高負荷判定(スケールアウト)の閾値です。
func (balancer LoadBalancer) ThHighInOperatingRatioAlgorithm(c *Config, w, n int) float64 {
	switch c.Cluster.LoadBalancer.OperatingRatioAlgorithm {
	case "BasicSpike":
		return balancer.OperatingRatioThresholdOut
	default:
		return Threshold
	}
}

// ThLowInOperatingRatioAlgorithm は稼働率ベースのアルゴリズムで使われる低負荷判定(スケールイン)の閾値です。
func (balancer LoadBalancer) ThLowInOperatingRatioAlgorithm(c *Config, w, n int) float64 {
	switch c.Cluster.LoadBalancer.OperatingRatioAlgorithm {
	case "BasicSpike":
		return balancer.OperatingRatioThresholdIn * float64(w)
	default:
		return balancer.OperatingRatioThresholdIn * float64(w)
		//return (Threshold - balancer.OperatingRatioDynamicThresholdDiff) * (float64(w))
	}
}

// Add
func (balancer LoadBalancer) Add(name string) error {
	// TODO haproxy setting
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
