package monitor

import "sync"

// Status
type Status struct {
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
	StatusCh          = make(chan Status, 1)
	PowerCh           = make(chan PowerStruct, 1)
	LoadCh            = make(chan float64, 1)
	DataCh            = make(chan []Data, 1)
	Statuses          []Status
	beforeTime        = map[string]int{}
	beforeTotalAccess = map[string]int{}
)

// GetStatusesはVMの名前，重さ，起動情報を保持する配列を返却します.
func GetStatuses() []Status {
	var mu sync.RWMutex
	mu.RLock()
	statuses := Statuses
	mu.RUnlock()

	return statuses
}