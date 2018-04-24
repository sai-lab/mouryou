package monitor

import "sync"

// State
type State struct {
	Name   string
	Weight int
	Info   string
}

// Data
type Data struct {
	Name       string
	Operating  float64
	Throughput int
	CPU        float64
}

type PowerStruct struct {
	Name string
	Info string
}

var (
	StateCh = make(chan State, 1)
	PowerCh = make(chan PowerStruct, 1)
	LoadCh  = make(chan float64, 1)
	DataCh  = make(chan []Data, 1)
	// 稼働状態
	States            []State
	beforeTime        = map[string]int{}
	beforeTotalAccess = map[string]int{}
)

// GetStatusesはVMの名前，重さ，起動情報を保持する配列を返却します.
func GetStates() []State {
	var mu sync.RWMutex
	mu.RLock()
	states := States
	mu.RUnlock()

	return states
}
