package monitor

import (
	"sync"
	"time"
)

// State
type State struct {
	Name   string
	Weight int
	// Infoはサーバの状態を示します。
	// 以下の値が入ります。
	// booting up は起動処理中を示します。
	// booted up は稼働中を示します。
	// shutting down は停止処理中を示します。
	// shutted down は停止中を示します。
	Info    string
	Changed time.Time
}

// Condition はサーバの状態を格納する構造体です．
// 稼働率，CPU利用率，タイムアウトなどのエラー情報を格納します．
type Condition struct {
	Name      string
	Operating float64
	CPU       float64
	Error     string
}

type PowerStruct struct {
	Name string
	Info string
	Load string
}

var (
	StateCh     = make(chan State, 1)
	PowerCh     = make(chan PowerStruct, 1)
	LoadORCh    = make(chan float64, 1)
	LoadTPCh    = make(chan float64, 1)
	ConditionCh = make(chan []Condition, 1)
	// 稼働状態
	States []State
)

// GetStatusesはVMの名前，重さ，起動情報を保持する配列を返却します.
func GetStates() []State {
	var mu sync.RWMutex
	mu.RLock()
	states := States
	mu.RUnlock()

	return states
}
