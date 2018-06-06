package engine

import (
	"container/ring"
	"fmt"
	"strconv"
	"sync"

	"github.com/sai-lab/mouryou/lib/convert"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
	"github.com/sai-lab/mouryou/lib/monitor"
	"github.com/sai-lab/mouryou/lib/mutex"
)

func Initialize(config *models.Config) {
	for name, machine := range config.Cluster.VirtualMachines {
		var st monitor.State
		st.Name = name
		if config.DevelopLogLevel >= 4 {
			logger.PrintPlace("Machine ID: " + strconv.Itoa(machine.Id) + ", Machine Name: " + name)
		}

		err := config.Cluster.LoadBalancer.ChangeWeight(name, machine.Weight)
		if err != nil {
			fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
			break
		}
		st.Weight = machine.Weight
		if config.DevelopLogLevel >= 4 {
			logger.PrintPlace("Machine ID: " + strconv.Itoa(machine.Id) + ", Machine Name: " + name)
		}

		if config.ContainID(machine.Id) {
			if config.DevelopLogLevel > 1 {
				logger.PrintPlace("LogLevel 1 : set booted up " + " Machine Name: " + name +
					" Weight: " + strconv.Itoa(machine.Weight))
			}
			st.Info = "booted up"
			totalWeight += machine.Weight
			futureTotalWeight += machine.Weight
		} else {
			st.Info = "shutted down"
			if config.DevelopLogLevel > 1 {
				logger.PrintPlace("LogLevel 1 : set shutted down " + " Machine Name: " + name +
					" Weight: " + strconv.Itoa(machine.Weight))
			}
		}
		monitor.States = append(monitor.States, st)
	}
}

func WeightOperator(config *models.Config) {
	var mu sync.RWMutex

	r := ring.New(LING_SIZE)

	mu.RLock()
	defer mu.RUnlock()

	for d := range monitor.DataCh {
		loadStates := map[string]int{}
		highLoads := map[string]int{}
		lowLoads := map[string]int{}
		weights := map[string]int{}
		cluster := config.Cluster
		weights["weights"] = -1

		for _, state := range monitor.States {
			if state.Name != "" {
				loadStates[state.Name] = 0
				if config.DevelopLogLevel >= 5 {
					logger.PrintPlace("state Name: " + state.Name + ", state weight: " + strconv.Itoa(state.Weight))
				}
				if state.Weight != 0 && state.Info == "booted up" {
					weights[state.Name] = state.Weight
				}
			}
		}

		r.Value = d
		r = r.Next()
		r.Do(func(v interface{}) {
			if v != nil {
				for _, ds := range v.([]monitor.Data) {
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
		}

		if len(highLoads) > 0 && len(lowLoads) > 0 {
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

func LoadCheck(ds monitor.Data, average int, threshold float64) int {
	loadState := 0

	if ds.Throughput != 0 {
		if ds.Throughput < average && ds.CPU <= 40 {
			// Throughput is low && Using CPU rate is low.
			loadState -= 1
		} else if (float64(ds.Throughput) < float64(average)*1.3 || float64(ds.Throughput) > float64(average)*0.8) && ds.CPU >= 90 {
			// At first glance, throuput is noting bad but using CPU rate is high.
			loadState += 1
		}
	}

	if ds.Operating >= threshold {
		// Operating rate is high.
		loadState += 1
	} else if ds.Operating <= 0.3 && ds.CPU <= 50 {
		// Operating rate is low && Using CPU rate is low.
		loadState -= 1
	}

	return loadState
}

func FireChangeWeight(config *models.Config, name string, w int) {
	var mu sync.RWMutex
	var err error

	mu.RLock()
	defer mu.RUnlock()
	for _, state := range monitor.States {
		if state.Name == name {
			if w < 0 && state.Weight <= 5 {
				fmt.Println(state.Name + " is low weight")
				break
			}
			s := monitor.State{state.Name, state.Weight, state.Info}
			s.Weight = state.Weight + w
			err = config.Cluster.LoadBalancer.ChangeWeight(s.Name, s.Weight)
			if err != nil {
				fmt.Println("Error is occured! Cannot change weight. Error is : " + fmt.Sprint(err))
				break
			}

			if monitor.StateCh != nil {
				monitor.StateCh <- s
			} else {
				fmt.Println("statusCh is nil")
			}
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+w)
			break
		}
	}
}
