package engine

import (
	"container/ring"
	"fmt"
	"sync"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/predictions"
	"github.com/sai-lab/mouryou/lib/logger"
)

//ServerManagement
func ServerManagement(c *models.Config) {
	var (
		totalOR        float64
		requiredNumber float64
		i              int
	    mu             sync.RWMutex
	)
	var b, w, s, tw, hw int
	var scaleIn bool

	r := ring.New(LING_SIZE)
	ttlORs := make([]float64, LING_SIZE)

	for totalOR = range monitor.LoadCh {
		r.Value = totalOR
		r = r.Next()
		ttlORs = convert.RingToArray(r)

		// Get Number of Active Servers
		w  = mutex.Read(&working, &workMutex)
		b  = mutex.Read(&booting, &bootMutex)
		s  = mutex.Read(&shuting, &shutMutex)
		tw = mutex.Read(&totalWeight, &totalWeightMutex)

		// Exec Algorithm
		if c.UseHetero {
			// Exec Algorithm for Server with Different Performace
			hw = predictions.ExecDifferentAlgorithm(c, w, b, s, tw, ttlORs)
			switch {
			case hw > tw:
				go BootUpVMs(c, hw-tw, requiredNumber)
			case hw < tw:
				go ShutDownVMs(c, tw-hw)
			}
		} else {
			// Exec Algorithm for Server with Same Performace
			requiredNumber, scaleIn = predictions.ExecSameAlgorithm(c, w, b, s, tw, ttlORs)
			mu.RLock()
			states := monitor.States
			mu.RUnlock()

			switch {
			case w+b < len(c.Cluster.VirtualMachines) && int(requiredNumber) > 0 && s == 0:
				for i = 0; i < int(requiredNumber); i++ {
					if w+b+i < len(c.Cluster.VirtualMachines) {
						for _, state := range states {
							if state.Info != "shutted down" {
								continue
							}
							go BootUpVM(c, state)
							mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+state.Weight)
						}
						logger.PrintPlace("Bootup")
					}
				}
			case w > 1 && scaleIn && mutex.Read(&waiting, &waitMutex) == 0 && b == 0:
				go ShutDownVMs(c, 10)
				fmt.Println("SM: Shutdown is fired")
			}
		}
	}
}

// BootUpVMs
func BootUpVMs(c *models.Config, weight int, requiredNumber float64) {
	var candidate []int
	var mu sync.RWMutex

	mu.RLock()
	states := monitor.States
	mu.RUnlock()

	for i, state := range states {
		if state.Info != "shutted down" {
			continue
		}
		if state.Weight >= weight {
			go BootUpVM(c, state)
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+state.Weight)
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
		go BootUpVM(c, states[boot])
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+states[boot].Weight)
	}
}

//BootUpVM
func BootUpVM(config *models.Config, st monitor.StatusStruct) {
	var p monitor.PowerStruct

	p.Name = st.Name
	p.Info = "booting up"
	st.Info = "booting up"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StatusCh != nil {
		monitor.StatusCh <- st
	}

	p.Info = config.Cluster.VirtualMachines[st.Name].Bootup(config.Sleep)
	st.Info = p.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StatusCh != nil {
		monitor.StatusCh <- st
	}
	fmt.Println(st.Name + "is booted up")
}

//ShutDownVMs
func ShutDownVMs(config *models.Config, weight int) {
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for _, st := range monitor.States {
		if st.Info != "booted up" {
			continue
		}
		if st.Weight <= weight {
			go ShutDownVM(config, st)
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-st.Weight)
			return
		}
	}
}

//ShutDownVM
func ShutDownVM(config *models.Config, st monitor.StatusStruct) {
	var p monitor.PowerStruct
	p.Name = st.Name
	p.Info = "shutting down"
	st.Info = "shutting down"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StatusCh != nil {
		monitor.StatusCh <- st
	}

	p.Info = config.Cluster.VirtualMachines[st.Name].Shutdown(config.Sleep)
	st.Info = p.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StatusCh != nil {
		monitor.StatusCh <- st
	}
}
