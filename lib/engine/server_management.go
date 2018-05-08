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
	"github.com/sai-lab/mouryou/lib/predictions"
)

// ServerManagement は起動状況と負荷状況に基いてオートスケールを実行します.
// 起動状況はengine.destination_settingが設定しています.
// 負荷状況はmonitor.LoadChから取得します.
func ServerManagement(c *models.Config) {
	var (
		// totalOR means the total value of the operating ratios of the working servers
		totalOR float64
		// w means the number of working servers
		w int
		// b means the number of booting servers
		b int
		// s means the number of servers that are stopped
		s int
		// tw means the total value of the weights of the working servers
		tw int
		// nw means the necessary weights
		nw int
	)

	r := ring.New(LING_SIZE)
	ttlORs := make([]float64, LING_SIZE)

	for totalOR = range monitor.LoadCh {
		r.Value = totalOR
		r = r.Next()
		ttlORs = convert.RingToArray(r)

		// Get Number of Active Servers
		w = mutex.Read(&working, &workMutex)
		b = mutex.Read(&booting, &bootMutex)
		s = mutex.Read(&shuting, &shutMutex)
		tw = mutex.Read(&totalWeight, &totalWeightMutex)

		// Exec Algorithm
		if c.UseHetero {
			nw = predictions.ExecDifferentAlgorithm(c, w, b, s, tw, ttlORs)
			switch {
			case nw > tw:
				go bootUpVMs(c, nw-tw)
			case nw < tw:
				go shutDownVMs(c, tw-nw)
			}
		} else {
			startStopSameServers(c, ttlORs, w, b, s, tw)
		}
	}
}

// startStopSameServersは単一性能向けアルゴリズムのサーバ起動停止メソッドです.
// predictions.ExecSameAlgorithmメソッドからmodels.Config.Sleep時間後に必要な台数と
// スケールインするかの真偽値を受け取り,それらに従って起動停止処理を実行します.
func startStopSameServers(c *models.Config, ttlORs []float64, w int, b int, s int, tw int) {
	var (
		scaleIn        bool
		requiredNumber float64
		i              int
	)

	requiredNumber, scaleIn = predictions.ExecSameAlgorithm(c, w, b, s, tw, ttlORs)
	statuses := monitor.GetStates()

	if c.DevelopLogLevel >= 3 {
		logger.PrintPlace("required server num is " + strconv.Itoa(int(requiredNumber)))
	}
	switch {
	case w+b < len(c.Cluster.VirtualMachines) && int(requiredNumber) > 0 && s == 0:
		for i = 0; i < int(requiredNumber); i++ {
			if w+b+i < len(c.Cluster.VirtualMachines) {
				for _, status := range statuses {
					if status.Info != "shutted down" {
						continue
					}
					go bootUpVM(c, status)
					mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+status.Weight)
					if c.DevelopLogLevel >= 1 {
						logger.PrintPlace("BootUp " + status.Name)
					}
				}
			}
		}
	case w > 1 && scaleIn && mutex.Read(&waiting, &waitMutex) == 0 && b == 0:
		go shutDownVMs(c, 10)
		if c.DevelopLogLevel >= 1 {
			fmt.Println("working number is " + strconv.Itoa(w))
			fmt.Println("SM: Shutdown is fired")
		}
	}
}

// bootUpVMs
func bootUpVMs(c *models.Config, weight int) {
	var candidate []int

	statuses := monitor.GetStates()

	for i, status := range statuses {
		// 停止中のサーバ以外は無視
		if status.Info != "shutted down" {
			continue
		}
		if status.Weight >= weight {
			go bootUpVM(c, status)
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+status.Weight)
			return
		}
		candidate = append(candidate, i)
	}

	if len(candidate) == 0 {
		return
	}

	boot := candidate[0]
	for _, n := range candidate {
		if statuses[n].Weight > statuses[boot].Weight {
			boot = n
		}
	}
	go bootUpVM(c, statuses[boot])
	mutex.Write(&totalWeight, &totalWeightMutex, totalWeight+statuses[boot].Weight)
}

// bootUpVM
func bootUpVM(c *models.Config, st monitor.State) {
	var p monitor.PowerStruct

	p.Name = st.Name
	p.Info = "booting up"
	st.Info = "booting up"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
	if c.DevelopLogLevel >= 1 {
		fmt.Println(st.Name + " is booting up")
	}

	p.Info = c.Cluster.VirtualMachines[st.Name].Bootup(c.Sleep)
	st.Info = p.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
	if c.DevelopLogLevel >= 1 {
		fmt.Println(st.Name + " is boot up")
	}
}

// shutDownVMs
func shutDownVMs(c *models.Config, weight int) {
	var mu sync.RWMutex

	mu.RLock()
	defer mu.RUnlock()

	for _, st := range monitor.States {
		// 稼働中のサーバ以外は無視
		if st.Info != "booted up" {
			continue
		}
		// オリジンサーバは無視
		if st.Name == c.OriginMachineName {
			continue
		}
		if st.Weight <= weight {
			go shutDownVM(c, st)
			mutex.Write(&totalWeight, &totalWeightMutex, totalWeight-st.Weight)
			if c.DevelopLogLevel >= 1 {
				fmt.Println(st.Name + " going to shutdown")
			}
			return
		}
	}
}

//shutDownVM
func shutDownVM(c *models.Config, st monitor.State) {
	var p monitor.PowerStruct
	p.Name = st.Name
	p.Info = "shutting down"
	st.Info = "shutting down"
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}

	p.Info = c.Cluster.VirtualMachines[st.Name].Shutdown(c.Sleep)
	st.Info = p.Info
	if monitor.PowerCh != nil {
		monitor.PowerCh <- p
	}
	if monitor.StateCh != nil {
		monitor.StateCh <- st
	}
}
