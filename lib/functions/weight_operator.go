package functions

import (
	"fmt"
	//"strings"
	"sync"

	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	//"github.com/sai-lab/mouryou/lib/timer"
)

func Initialize(config *models.ConfigStruct) {
	i := 0 // debug
	for name, machine := range config.Cluster.VirtualMachines {
		var st StatusStruct
		st.Name = name
		st.Weight = 1

		if machine.Id == 1 {
			st.Info = "booted up"
			states = append(states, st)
			logger.PrintPlace("machine.Name = " + machine.Name + ", state.Name = " + states[i].Name + ", state.Weight = " + fmt.Sprint(states[i].Weight) + ", state.Info = " + states[i].Info) // debug
			i++                                                                                                                                                                                //debug
			continue
		}

		st.Info = "shutted down"
		states = append(states, st)
		logger.PrintPlace("machine.Name = " + machine.Name + ", state.Name = " + states[i].Name + ", state.Weight = " + fmt.Sprint(states[i].Weight) + ", state.Info = " + states[i].Info) //debug
		i++                                                                                                                                                                                // debug
	}
}

func BootUpVMs(config *models.ConfigStruct, weight float64) {
	var candidate []int
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for i, state := range states {
		logger.PrintPlace("check states, st.Name: " + state.Name + ", st.Weight: " + fmt.Sprint(state.Weight) + ", st.Info: " + state.Info)
		if state.Info != "shutted down" {
			continue
		}
		if state.Weight >= weight {
			go BootUpVM(config, state)
			mutex.WriteFloat(&totalWeight, &totalWeightMutex, totalWeight+state.Weight)
			return
		} else {
			logger.PrintPlace("candidate add" + state.Name)
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
	logger.PrintPlace("Booting Up, st.Name: " + st.Name)

	p.Info = config.Cluster.VirtualMachines[st.Name].Bootup(config.Sleep)
	st.Info = p.Info
	if powerCh != nil {
		powerCh <- p
	}
	if statusCh != nil {
		logger.PrintPlace("channel is sended")
		statusCh <- st
	}
	logger.PrintPlace("Booted Up, st.Name: " + st.Name)
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

func ChangeWeight(hostname string, state string) {
	switch state {
	case "critical":
	case "light":
	}

}
