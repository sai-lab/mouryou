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
	UseWeb          bool `json:"use_web"`
	DevelopLogLevel int  `json:"develop_log_level"`
	// 負荷取得がタイムアウトしたと判断するまでの時間
	Timeout time.Duration `json:"timeout"`
	Sleep   time.Duration `json:"sleep"`
	Wait    time.Duration `json:"wait"`
	// タイムアウトなどして重さを下げた後、復元するまでの時間
	RestorationTime time.Duration `json:"restoration_time"`
	// 起動までのMargin
	Margin float64 `json:"margin"`
	// 利用するアルゴリズム
	Algorithm string `json:"algorithm"`
	// 重さを変更するか
	IsWeightChange bool `json:"is_weight_change"`
	// ヘテロな環境を使用するか
	UseHetero                   bool            `json:"use_hetero"`
	AdjustServerNum             bool            `json:"adjust_server_num"`
	OriginMachineNames          []string        `json:"origin_machine_names"`
	AlwaysRunningMachines       []string        `json:"always_running_machines"`
	StartMachineIDs             []int           `json:"start_machine_ids"`
	WebSocket                   WebSocketStruct `json:"web_socket"`
	Cluster                     Cluster         `json:"cluster"`
	UseOperatingRatio           bool            `json:"use_operating_ratio"`
	UseThroughput               bool            `json:"use_throughput"`
	ThroughputAlgorithm         string          `json:"throughput_algorithm"`
	ThroughputScaleOutThreshold int             `json:"throughput_scale_out_threshold"`
	ThroughputScaleInThreshold  int             `json:"throughput_scale_in_threshold"`
	ThroughputScaleInRate       float64         `json:"throughput_scale_in_rate"`
	ThroughputScaleOutTime      int             `json:"throughput_scale_out_time"`
	ThroughputScaleInTime       int             `json:"throughput_scale_in_time"`
	LogDB                       string          `json:"log_db"`
	LogDSN                      string          `json:"log_dsn"`
	InfluxDBAddr                string          `json:"influxdb_addr"`
	InfluxDBPort                string          `json:"influxdb_port"`
	InfluxDBUser                string          `json:"influxdb_user"`
	InfluxDBPasswd              string          `json:"influxdb_passwd"`
	InfluxDBConnection          client.Client   `json:"influxdb_connection"`
	InfluxDBServerDB            string          `json:"influxdb_serverdb"`
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

// ContainMachineName はnameがnamesに含まれているかどうか検証します。
func (*Config) ContainMachineName(names []string, name string) bool {
	for _, n := range names {
		if name == n {
			return true
		}
	}
	return false
}
