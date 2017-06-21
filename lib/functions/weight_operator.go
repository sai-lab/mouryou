package functions

import (
	"fmt"
	//"strings"
	"container/ring"
	"sync"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
	//"github.com/sai-lab/mouryou/lib/timer"
)

func Initialize(config *models.ConfigStruct) {
	for name, machine := range config.Cluster.VirtualMachines {
		var st StatusStruct
		st.Name = name

		if machine.Id == 1 || machine.Id == 3 || machine.Id == 5 || machine.Id == 8 || machine.Id == 9 {
			st.Info = "booted up"
			st.Weight = machine.Weight
			states = append(states, st)
			totalWeight += machine.Weight
			continue
		}

		st.Info = "shutted down"
		states = append(states, st)
	}
}

func WeightOperator(config *models.ConfigStruct) {
	var mu sync.RWMutex
	var th, prevThroughPut float64

	r := ring.New(LING_SIZE)

	mu.RLock()
	defer mu.RUnlock()

	for d := range dataCh {
		cs := map[string]int{}
		highCs := map[string]int{}
		lowCs := map[string]int{}
		weights := map[string]int{}
		cluster := config.Cluster
		prevThroughPut = 0
		weights["weights"] = -1

		for _, state := range states {
			if state.Name != "" {
				cs[state.Name] = 0
				if state.Weight != 0 {
					weights[state.Name] = state.Weight
				}
			}
		}

		r.Value = d
		r = r.Next()
		r.Do(func(v interface{}) {
			if v != nil {
				th = 0
				for _, ds := range v.([]DataStruct) {
					if prevThroughPut < 1 {
						th = -1
					} else {
						th = ds.ThroughPut - prevThroughPut
					}
					if ds.Operating >= models.Threshold {
						cs[ds.Name] += 1
					} else if ds.Operating <= 0.3 && ds.Cpu <= 50 {
						cs[ds.Name] -= 1
					} else if ds.ThroughPut < float64(cluster.VirtualMachines[ds.Name].Average)/2 {
						cs[ds.Name] -= 1
					}
					if th != -1 {
						if ds.ThroughPut < cluster.VirtualMachines[ds.Name].Average && ds.Cpu <= 40 {
							cs[ds.Name] -= 1
						} else if (ds.ThroughPut < cluster.VirtualMachines[ds.Name].Average*1.3 || ds.ThroughPut > cluster.VirtualMachines[ds.Name].Average*0.8) && ds.Cpu >= 80 {
							cs[ds.Name] += 1
						}
					}
				}
			}
		})
		for k, v := range cs {
			if v < -5 {
				lowCs[k] = v
			} else if v > 5 {
				highCs[k] = v
			}
		}
		if len(highCs) > 0 && len(lowCs) > 0 {
			fmt.Println("highCs: " + fmt.Sprint(highCs))
			fmt.Println("lowCs: " + fmt.Sprint(lowCs))
			for name, _ := range lowCs {
				FireChangeWeight(config, name, 5)
				weights[name] = weights[name] + 5
			}
			for name, _ := range highCs {
				if weights[name] <= 5 {
					continue
				}
				FireChangeWeight(config, name, -5)
				weights[name] = weights[name] - 5
			}
		}
		ar := convert.MapToArray(weights)
		logger.Write(ar)
		logger.Print(ar)
	}
}

func FireChangeWeight(config *models.ConfigStruct, name string, w int) {
	var mu sync.RWMutex
	var err error

	mu.RLock()
	defer mu.RUnlock()
	for _, state := range states {
		if state.Name == name {
			if w < 0 && state.Weight <= 5 {
				fmt.Println(state.Name + " is low weight")
				break
			}
			s := StatusStruct{state.Name, state.Weight, state.Info}
			s.Weight = state.Weight + w
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
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+w)
			break
		}
	}
}
