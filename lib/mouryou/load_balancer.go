package mouryou

import (
	"os/exec"
)

type LoadBalancerStruct struct {
	VirtualIP string  `json:"virtual_ip"`
	Algorithm string  `json:"algorithm"`
	Threshold float64 `json:"threshold"`
	Margin    float64 `json:"margin"`
	ScaleOut  int     `json:"scale_out"`
	ScaleIn   int     `json:"scale_in"`
}

func (balancer LoadBalancerStruct) Initialize() {
	exec.Command("ip", "addr", "add", balancer.VirtualIP, "label", "eth0:vip", "dev", "eth0").Run()

	err := exec.Command("ipvsadm", "-C").Run()
	checkError(err)

	err = exec.Command("ipvsadm", "-A", "-t", balancer.VirtualIP+":http", "-s", balancer.Algorithm).Run()
	checkError(err)
}

func (balancer LoadBalancerStruct) ThHigh() float64 {
	return balancer.Threshold
}

func (balancer LoadBalancerStruct) ThLow(n int) float64 {
	return balancer.ThHigh()*float64(n-1)/float64(n) - balancer.Margin
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
