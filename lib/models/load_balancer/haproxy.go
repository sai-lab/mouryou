package models

import (
	"os/exec"

	"github.com/sai-lab/mouryou/lib/check"
)

type HaproxyStruct struct {
	LoadBalancerStruct
}

func (balancer LoadBalancerStruct) Initialize() {
	err := exec.Command("ipvsadm", "-C").Run()
	check.Error(err)

	err = exec.Command("ipvsadm", "-A", "-t", balancer.VirtualIP+":http", "-s", balancer.Algorithm).Run()
	check.Error(err)
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
