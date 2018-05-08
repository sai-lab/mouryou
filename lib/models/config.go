package models

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"errors"
	"fmt"
	"reflect"

	"github.com/sai-lab/mouryou/lib/check"
)

type Config struct {
	// DevelopLogLevel はデバッグ用標準出力のレベルを指定します。
	// 必要ないと判断したら消してください。
	// DevelopLogLevel==0: デフォルト 標準出力なし
	// DevelopLogLevel>=1: 起動停止操作等を出力
	// DevelopLogLevel>=2: 各サーバの重み情報を出力
	// DevelopLogLevel>=3: 各サーバの負荷状況を全て出力
	// DevelopLogLevel>=4: 詳細に
	DevelopLogLevel   int             `json:"develop_log_level"`
	Timeout           time.Duration   `json:"timeout"`
	Sleep             time.Duration   `json:"sleep"`
	Wait              time.Duration   `json:"wait"`
	Margin            float64         `json:"margin"`
	Algorithm         string          `json:"algorithm"`
	UseHetero         bool            `json:"use_hetero"`
	AdjustServerNum   bool            `json:"adjust_server_num"`
	OriginMachineName string          `json:"origin_machine_name"`
	StartMachineIDs   []int           `json:"start_machine_ids"`
	WebSocket         WebSocketStruct `json:"web_socket"`
	Cluster           Cluster         `json:"cluster"`
}

// LoadConfig は設定ファイル(~/.mouryou.json)を読み込みます。
// 設定されていない値があるとここで処理を終了します。
func (c *Config) LoadSetting(path string) {
	bytes, err := ioutil.ReadFile(path)
	check.Error(err)

	err = json.Unmarshal(bytes, &c)
	check.Error(err)

	err = c.valueCheck()
	check.Error(err)

	Threshold = c.Cluster.LoadBalancer.ThresholdOut
}

// valueCheck は設定ファイル(~/.mouryou.json)に各設定値が記述されているかチェックします。
// 記述がない設定値があるとerrorを返します。
func (c *Config) valueCheck() error {
	var err error
	var errTxt string
	var e string
	intNil := []int(nil)

	if c.Timeout == time.Duration(0) {
		e = "please set timeout for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + e
	}
	if c.Sleep == time.Duration(0) {
		e := "please set sleep for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if c.Wait == time.Duration(0) {
		e := "please set wait for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if c.Margin == float64(0) {
		e := "please set margin for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if c.Algorithm == "" {
		e := "please set algorithm for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if c.OriginMachineName == "" {
		e := "please set originMachineName for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if reflect.DeepEqual(c.StartMachineIDs, intNil) {
		e := "please set startMachineIDs for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}

	if errTxt != "" {
		err = errors.New(errTxt)
	}

	return err
}

// ContainID は受け取ったVMのIDが開始時から稼動状態とするサーバに
// 指定されているかどうか検証します。
func (c *Config) ContainID(i int) bool {
	for _, v := range c.StartMachineIDs {
		if i == v {
			return true
		}
	}
	return false
}
