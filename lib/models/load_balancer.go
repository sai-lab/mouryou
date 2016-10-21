package models

import (
	"os/exec"

	"github.com/sai-lab/mouryou/lib/check"
)

type LoadBalancerStruct struct {
	VirtualIP     string  `json:"virtual_ip"`
	Algorithm     string  `json:"algorithm"`
	Threshold_out float64 `json:"threshold_out"`
	Threshold_in  float64 `json:"threshold_in"`
	Margin        float64 `json:"margin"`
	ScaleOut      int     `json:"scale_out"`
	ScaleIn       int     `json:"scale_in"`
}

func (balancer LoadBalancerStruct) Initialize() {
	exec.Command("ip", "addr", "add", balancer.VirtualIP, "label", "eth0:vip", "dev", "eth0").Run()

	err := exec.Command("ipvsadm", "-C").Run()
	check.Error(err)

	err = exec.Command("ipvsadm", "-A", "-t", balancer.VirtualIP+":http", "-s", balancer.Algorithm).Run()
	check.Error(err)
}

func (balancer LoadBalancerStruct) ThHigh(w, n int) float64 {
	return balancer.Threshold_out
}

func (balancer LoadBalancerStruct) ThLow(w int) float64 {
	return balancer.Threshold_in
}

func (balancer LoadBalancerStruct) Add(host string) error {
	err := exec.Command("ipvsadm", "-a", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "0", "-g").Run()
	if err != nil {
		return err
	}

	return nil
}

func (balancer LoadBalancerStruct) Remove(host string) error {
	err := exec.Command("ipvsadm", "-d", "-t", balancer.VirtualIP+":http", "-r", host+":http").Run()
	if err != nil {
		return err
	}

	return nil
}

func (balancer LoadBalancerStruct) Active(host string) error {
	err := exec.Command("ipvsadm", "-e", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "1", "-g").Run()
	if err != nil {
		return err
	}

	return nil
}

func (balancer LoadBalancerStruct) Inactive(host string) error {
	err := exec.Command("ipvsadm", "-e", "-t", balancer.VirtualIP+":http", "-r", host+":http", "-w", "0", "-g").Run()
	if err != nil {
		return err
	}

	return nil
}
