package engine

import (
	"sync"
)

//LingSize は合計稼働率totalORを保持するLINGのサイズです．
const LingSize = 10

// autoScaleOrder はオートスケール命令を格納する構造体です．
// load_determination.goでscaleChに送信され，
// server_management.goで受信されます．
type autoScaleOrder struct {
	Handle string
	Weight int
	Load   string
}

var (
	working           = 1 // 稼働中の台数
	booting           = 0 // 起動処理中の台数
	shutting          = 0 // 停止処理中の台数
	waiting           = 0 // 停止処理を待つフラグ
	waits             = 0 // 停止処理待ちの台数
	totalWeight       = 0 // 稼働中のサーバの重みの合計値
	futureTotalWeight = 0 // サーバの起動・停止処理完了後の重みの合計値

	workMutex              sync.RWMutex // workingへの読み書き排他制御
	bootMutex              sync.RWMutex // bootingへの読み書き排他制御
	shutMutex              sync.RWMutex // shuttingへの読み書き排他制御
	waitMutex              sync.RWMutex // waitingへの読み書き排他制御
	waitsMutex             sync.RWMutex // waitsへの読み書き排他制御
	totalWeightMutex       sync.RWMutex // totalWeightへの読み書き排他制御
	futureTotalWeightMutex sync.RWMutex // futureTotalWeightへの読み書き排他制御

	autoScaleOrderCh = make(chan autoScaleOrder, 1) //オートスケール命令チャネル
)
