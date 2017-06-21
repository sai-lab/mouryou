package functions

import (
	"container/ring"
	"fmt"
	"sync"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/mutex"
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

	r := ring.New(LING_SIZE)

	mu.RLock()
	defer mu.RUnlock()

	for d := range dataCh {
		loadStates := map[string]int{}
		highLoads := map[string]int{}
		lowLoads := map[string]int{}
		weights := map[string]int{}
		cluster := config.Cluster
		weights["weights"] = -1

		for _, state := range states {
			if state.Name != "" {
				loadStates[state.Name] = 0
				if state.Weight != 0 {
					weights[state.Name] = state.Weight
				}
			}
		}

		r.Value = d
		r = r.Next()
		r.Do(func(v interface{}) {
			if v != nil {
				for _, ds := range v.([]DataStruct) {
					loadStates[ds.Name] += LoadCheck(ds, cluster.VirtualMachines[ds.Name].Average, models.Threshold)
				}
			}
		})

		// check server load
		for k, v := range loadStates {
			if v < -5 {
				lowLoads[k] = v
			} else if v > 5 {
				highLoads[k] = v
			}
			// fmt.Println(k + " " + strconv.Itoa(v))
		}

		if len(highLoads) > 0 && len(lowLoads) > 0 {
			fmt.Println("highLoads: " + fmt.Sprint(highLoads))
			fmt.Println("lowLoads: " + fmt.Sprint(lowLoads))

			for name, _ := range lowLoads {
				FireChangeWeight(config, name, 5)
				weights[name] = weights[name] + 5
			}
			for name, _ := range highLoads {
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

func LoadCheck(ds DataStruct, average int, threshold float64) int {
	loadState := 0

	if ds.Throughput != 0 {
		if ds.Throughput < average && ds.Cpu <= 40 {
			// Throughput is low && Using CPU rate is low.
			loadState -= 1
		} else if (float64(ds.Throughput) < float64(average)*1.3 || float64(ds.Throughput) > float64(average)*0.8) && ds.Cpu >= 90 {
			// At first glance, throuput is noting bad but using CPU rate is high.
			loadState += 1
		}
	}

	if ds.Operating >= threshold {
		// Operating rate is high.
		loadState += 1
	} else if ds.Operating <= 0.3 && ds.Cpu <= 50 {
		// Operating rate is low && Using CPU rate is low.
		loadState -= 1
	}

	return loadState
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
