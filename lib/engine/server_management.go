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
)

func ServerManagement(c *models.ConfigStruct) {
	var ttlOR, n float64
	var b, w, s, tw int
	var scaleIn bool

	r := ring.New(LING_SIZE)
	ttlORs := make([]float64, LING_SIZE)

	for ttlOR = range monitor.LoadCh {
		r.Value = ttlOR
		r = r.Next()
		ttlORs = convert.RingToArray(r)

		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shuting, &shutMutex)
		tw = mutex.Read(&totalWeight, &totalWeightMutex)

		// Exec Algorithm
		n, scaleIn = predictions.Exec(c, w, b, s, tw, ttlORs)

		// --- Periodically Prediction Algorithm
		hw := predictions.PeriodicallyPrediction(w, b, s, tw)
		switch {
		case hw > tw:
			// go BootUpVMs(config, hw-tw)
		case hw < tw:
			// go ShutDownVMs(config, tw-hw)
		}
		/// ---

		// --- Basic Spike Prediction Algorithm's Server Management
		switch {
		case int(n) > tw && int(n) > 0 && s == 0:
			if w+b < len(c.Cluster.VirtualMachines) {
				// go BootUpVMs(config, n-tw)
				// fmt.Println("SM: BootUp is fired. n: " + fmt.Sprintf("%3.5f", n) + ", tw: " + fmt.Sprintf("%3.5f", tw))
			}
		case w > 1 && scaleIn && mutex.Read(&waiting, &waitMutex) == 0 && b == 0:
			// go ShutDownVMs(config, 10)
			// fmt.Println("SM: Shutdown is fired")
		}
		// ---
	}
}

func BootUpVMs(config *models.ConfigStruct, weight int) {
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
			go BootUpVM(config, state)
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
		go BootUpVM(config, states[boot])
		mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+states[boot].Weight)
	}
}

func BootUpVM(config *models.ConfigStruct, st monitor.StatusStruct) {
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

func ShutDownVMs(config *models.ConfigStruct, weight int) {
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

func ShutDownVM(config *models.ConfigStruct, st monitor.StatusStruct) {
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
