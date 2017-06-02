package functions

import (
	"fmt"
	//"strings"
	"sync"

	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	//"github.com/sai-lab/mouryou/lib/timer"
)

func Initialize(config *models.ConfigStruct) {
	for name, machine := range config.Cluster.VirtualMachines {
		var st StatusStruct
		st.Name = name
		st.Weight = 10

		if machine.Id == 1 {
			st.Info = "booted up"
			states = append(states, st)
			continue
		}

		st.Info = "shutted down"
		states = append(states, st)
	}
}

func BootUpVMs(config *models.ConfigStruct, weight float64) {
	var candidate []int
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for i, state := range states {
		if state.Info != "shutted down" {
			continue
		}
		if state.Weight >= weight {
			go BootUpVM(config, state)
			mutex.WriteFloat(&totalWeight, &totalWeightMutex, totalWeight+state.Weight)
			return
		} else {
			candidate = append(candidate, i)
		}
	}

	if len(candidate) == 0 {
		return
	} else {
		boot := candidate[0]
		for _, n := range candidate {
			if states[n].Weight > states[boot].Weight {
				boot = n
			}
		}
		go BootUpVM(config, states[boot])
		mutex.WriteFloat(&totalWeight, &totalWeightMutex, totalWeight+states[boot].Weight)
	}
}

func BootUpVM(config *models.ConfigStruct, st StatusStruct) {
	var p PowerStruct

	p.Name = st.Name
	p.Info = "booting up"
	st.Info = "booting up"
	if powerCh != nil {
		powerCh <- p
	}
	if statusCh != nil {
		statusCh <- st
	}

	p.Info = config.Cluster.VirtualMachines[st.Name].Bootup(config.Sleep)
	st.Info = p.Info
	if powerCh != nil {
		powerCh <- p
	}
	if statusCh != nil {
		statusCh <- st
	}
	fmt.Println(st.Name + "is booted up")
}

func ShutDownVMs(config *models.ConfigStruct, weight float64) {
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for _, st := range states {
		if st.Info != "booted up" {
			continue
		}
		if st.Weight <= weight {
			go ShutDownVM(config, st)
			mutex.WriteFloat(&totalWeight, &totalWeightMutex, totalWeight-st.Weight)
			return
		}
	}
}

func ShutDownVM(config *models.ConfigStruct, st StatusStruct) {
	var p PowerStruct
	p.Name = st.Name
	p.Info = "shutting down"
	st.Info = "shutting down"
	if powerCh != nil {
		powerCh <- p
	}
	if statusCh != nil {
		statusCh <- st
	}

	p.Info = config.Cluster.VirtualMachines[st.Name].Shutdown(config.Sleep)
	st.Info = p.Info
	if powerCh != nil {
		powerCh <- p
	}
	if statusCh != nil {
		statusCh <- st
	}
}

func MonitorWeightChange(config *models.ConfigStruct) {
	var cr CriticalStruct

	for cr = range criticalCh {
		fmt.Println("get critical message")
		switch cr.Info {
		case "critical":
			go FireChangeWeight(config, cr, 1, false)
		case "light":
			go FireChangeWeight(config, cr, 1, true)
		default:
		}
	}

	fmt.Println("MonitorWeightChange is finished!")
}

func FireChangeWeight(config *models.ConfigStruct, cr CriticalStruct, w float64, incOrDec bool) {
	var mu sync.RWMutex
	var err error

	mu.RLock()
	defer mu.RUnlock()
	for _, state := range states {
		if state.Name == cr.Name {
			if !incOrDec && state.Weight <= 5 {
				fmt.Println(state.Name + " is low weight")
				break
			}
			s := StatusStruct{state.Name, state.Weight, state.Info}
			if incOrDec {
				s.Weight = state.Weight + w
			} else {
				s.Weight = state.Weight - w
			}

			err = config.Cluster.LoadBalancer.ChangeWeight(s.Name, s.Weight)
			if err != nil {
				fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
				break
			}
			if statusCh != nil {
				statusCh <- s
			} else {
				fmt.Println("statusCh is nil")
			}
			mutex.WriteFloat(&totalWeight, &totalWeightMutex, totalWeight-1)
			break
		}
	}
}
