package functions

import (
	"container/ring"
	"fmt"
	"sync"

	"github.com/sai-lab/mouryou/lib/calculate"
	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	"github.com/sai-lab/mouryou/lib/ratio"
)

func ServerManagement(config *models.ConfigStruct) {
	var ttlor, out, in, ThHigh, ThLow, ir, n float64
	var b, w, s, tw int

	r := ring.New(LING_SIZE)
	ttlors := make([]float64, LING_SIZE)

	for ttlor = range loadCh {
		r.Value = ttlor
		r = r.Next()

		ttlors = convert.RingToArray(r)
		out = calculate.MovingAverage(ttlors, config.Cluster.LoadBalancer.ScaleOut)
		in = calculate.MovingAverage(ttlors, config.Cluster.LoadBalancer.ScaleIn)

		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shuting, &shutMutex)
		tw = mutex.Read(&totalWeight, &totalWeightMutex)
		// config.Cluster.LoadBalancer.ChangeThresholdOut(w, b, s, len(config.Cluster.VirtualMachines))
		ThHigh = config.Cluster.LoadBalancer.ThHigh(w, len(config.Cluster.VirtualMachines))
		ThLow = config.Cluster.LoadBalancer.ThLow(w)

		ir = ratio.Increase(ttlors, config.Cluster.LoadBalancer.ScaleOut)
		n = (((out + ir*float64(config.Sleep)) / ThHigh) - float64(w+b)) * 10
		weights := []string{"we", fmt.Sprintf("%3.5f", n), fmt.Sprintf("%3d", tw)}
		logger.Print(weights)
		logger.Write(weights)

		switch {
		case int(n) > tw && int(n) > 0 && s == 0:
			if w+b < len(config.Cluster.VirtualMachines) {
				// go BootUpVMs(config, n-tw)
				// fmt.Println("SM: BootUp is fired. n: " + fmt.Sprintf("%3.5f", n) + ", tw: " + fmt.Sprintf("%3.5f", tw))
			}
		case w > 1 && in < ThLow && mutex.Read(&waiting, &waitMutex) == 0 && b == 0:
			// go ShutDownVMs(config, 10)
			// fmt.Println("SM: Shutdown is fired")
		}
	}
}

func BootUpVMs(config *models.ConfigStruct, weight int) {
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

func ShutDownVMs(config *models.ConfigStruct, weight int) {
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for _, st := range states {
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
