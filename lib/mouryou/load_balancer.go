package mouryou

import (
	"os/exec"
)

type loadBalancer struct {
	VIP        string
	Algorithem string
	ThHigh     float64 `json:th_high`
	ScaleOut   int     `json:scale_out`
	ScaleIn    int     `json:scale_in`
}

func (lb loadBalancer) thHigh() float64 {
	return lb.ThHigh
}

func (lb loadBalancer) thLow(working int) float64 {
	return lb.thHigh()*float64(working-1)/float64(working) - 0.05
}

func (lb loadBalancer) init() {
	exec.Command("ip", "addr", "add", lb.VIP, "label", "eth0:vip", "dev", "eth0").Run()

	err := exec.Command("ipvsadm", "-C").Run()
	checkError(err)

	err = exec.Command("ipvsadm", "-A", "-t", lb.VIP+":http", "-s", lb.Algorithem).Run()
	checkError(err)
}

func (lb loadBalancer) add(real string) {
	err := exec.Command("ipvsadm", "-a", "-t", lb.VIP+":http", "-r", real+":http", "-w", "0", "-g").Run()
	checkError(err)
}

func (lb loadBalancer) active(real string) {
	err := exec.Command("ipvsadm", "-e", "-t", lb.VIP+":http", "-r", real+":http", "-w", "1", "-g").Run()
	checkError(err)
}

func (lb loadBalancer) inactive(real string) {
	err := exec.Command("ipvsadm", "-e", "-t", lb.VIP+":http", "-r", real+":http", "-w", "0", "-g").Run()
	checkError(err)
}
