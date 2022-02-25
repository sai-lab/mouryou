package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/influxdata/influxdb/client/v2"
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
	DevelopLogLevel       int             `json:"develop_log_level"`
	Timeout               time.Duration   `json:"timeout"` // 負荷取得がタイムアウトしたと判断するまでの時間
	Sleep                 time.Duration   `json:"sleep"`   // サーバの起動処理発行後、稼働し始めるまでの時間
	Start                 time.Duration   `json:"start"`
	Stop                  time.Duration   `json:"stop"`
	Wait                  time.Duration   `json:"wait"`                    // 起動処理発行後、停止処理を実行しない時間
	RestorationTime       time.Duration   `json:"restoration_time"`        // タイムアウトなどして重さを下げた後、復元するまでの時間
	IsWeightChange        bool            `json:"is_weight_change"`        // 重さを変更するか
	UseHetero             bool            `json:"use_hetero"`              // ヘテロな環境を使用するか
	IsAdjustServerNum     bool            `json:"is_adjust_server_num"`    // オートスケールを行うか(現状未使用)
	UseWeb                bool            `json:"use_web"`                 // mouryou-webに情報を送るか
	UseOperatingRatio     bool            `json:"use_operating_ratio"`     // 稼働率ベースの負荷判定アルゴリズムを使うか
	UseThroughput         bool            `json:"use_throughput"`          // スループットベースの負荷判定アルゴリズムを使うか
	InfluxDBAddr          string          `json:"influxdb_addr"`           // InfluxDBのアドレス
	InfluxDBPort          string          `json:"influxdb_port"`           // InfluxDBのポート
	InfluxDBUser          string          `json:"influxdb_user"`           // InfluxDBのユーザ
	InfluxDBPasswd        string          `json:"influxdb_passwd"`         // InfluxDBのパスワード
	InfluxDBConnection    client.Client   `json:"influxdb_connection"`     // InfluxDBへのコネクション
	InfluxDBServerDB      string          `json:"influxdb_serverdb"`       // InfluxDBで利用するDB名
	OriginMachineNames    []string        `json:"origin_machine_names"`    // オリジンサーバの名称
	AlwaysRunningMachines []string        `json:"always_running_machines"` // 常に稼働するサーバの名称
	StartMachineIDs       []int           `json:"start_machine_ids"`       // はじめから稼働するサーバのID
	WebSocket             WebSocketStruct `json:"web_socket"`              // models.WebSocketの構造体
	Cluster               Cluster         `json:"cluster"`                 // models.Clusterの構造体
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

	Threshold = c.Cluster.LoadBalancer.OperatingRatioThresholdOut
}

// valueCheck は設定ファイル(~/.mouryou.json)に各設定値が記述されているかチェックします。
// 記述がない設定値があるとerrorを返します。
func (c *Config) valueCheck() error {
	var err error
	var errTxt string
	var e string
	stringNil := []string(nil)
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
	if reflect.DeepEqual(c.OriginMachineNames, stringNil) {
		e := "please set origin_machine_names for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if reflect.DeepEqual(c.AlwaysRunningMachines, stringNil) {
		e := "please set always_running_machines for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}
	if reflect.DeepEqual(c.StartMachineIDs, intNil) {
		e := "please set start_machine_ids for mouryou.json"
		fmt.Println(e)
		errTxt = errTxt + ", " + e
	}

	if errTxt != "" {
		err = errors.New(errTxt)
	}

	return err
}

// IsStartMachineID は受け取ったVMのIDが開始時から稼動状態とするサーバに
// 指定されているかどうか検証します。
func (c *Config) IsStartMachineID(i int) bool {
	for _, v := range c.StartMachineIDs {
		if i == v {
			return true
		}
	}
	return false
}

// ContainMachineName はnameがnamesに含まれているかどうか検証します。
func (*Config) ContainMachineName(names []string, name string) bool {
	for _, n := range names {
		if name == n {
			return true
		}
	}
	return false
}
