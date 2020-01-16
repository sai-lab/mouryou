package monitor

import (
	"errors"
	"sync"
	"time"
)

// ServerState はサーバの名前，重み，状態，変更を保持します．
type ServerState struct {
	Name   string
	Weight int
	// Infoはサーバの状態を示します。
	// 以下の値が入ります。
	// booting up は起動処理中を示します。
	// booted up は稼働中を示します。
	// shutting down は停止処理中を示します。
	// shutted down は停止中を示します。
	Info     string
	Changed  time.Time
	WaitTime time.Time
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
	// StateCh は，engine/server_management.goから送信され，
	// engine.StatusManager()で受信されます．
	StateCh  = make(chan ServerState, 1)
	PowerCh  = make(chan PowerStruct, 1)
	LoadORCh = make(chan float64, 1)
	LoadTPCh = make(chan float64, 1)
	// ConditionCh は，engine.Ratios()から送信され，
	// engine.WeightManager()で受信されます．
	ConditionCh = make(chan []Condition, 1)
	// TODO ローカル変数にしたい
	ServerStates []ServerState
)

// AddServerState は ServerStates に新しい要素を追加します．
func AddServerState(state ServerState) error {
	var mu sync.RWMutex
	mu.RLock()
	ServerStates = append(ServerStates, state)
	mu.RUnlock()

	return nil
}

// GetServerStates はVMの名前，重さ，起動情報を保持する配列を返却します.
func GetServerStates() []ServerState {
	var mu sync.RWMutex
	mu.RLock()
	serverStates := ServerStates
	mu.RUnlock()

	return serverStates
}

// UpdateServerStates は ServerState の情報を更新します．
func UpdateServerStates(hostName string, weight int, info string, changed time.Time, wait time.Time) error {

	isUpdated := false
	for i, state := range ServerStates {
		if state.Name == hostName {
			if weight >= 0 {
				// 0以上ならば有効な値と仮定
				ServerStates[i].Weight = weight
			}
			if info != "" {
				// 空文字でなければ有効な値と仮定
				ServerStates[i].Info = info
			}
			if changed.IsZero() == false {
				ServerStates[i].Changed = changed
			}
			if wait.IsZero() == false {
				ServerStates[i].WaitTime = wait
			}
			isUpdated = true
			break
		}
	}

	if !isUpdated {
		return errors.New("hostName is not found")
	}

	return nil
}
