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
	DevelopLogLevel int             `json:"develop_log_level"`
	Timeout         time.Duration   `json:"timeout"`
	Sleep           time.Duration   `json:"sleep"`
	Wait            time.Duration   `json:"wait"`
	Margin          float64         `json:"margin"`
	Algorithm       string          `json:"algorithm"`
	UseHetero       bool            `json:"use_hetero"`
	AdjustServerNum bool            `json:"adjust_server_num"`
	StartMachineIDs []int           `json:"start_machine_ids"`
	WebSocket       WebSocketStruct `json:"web_socket"`
	Cluster         Cluster         `json:"cluster"`
}

//LoadConfig
func (c *Config) LoadSetting(path string) {
	bytes, err := ioutil.ReadFile(path)
	check.Error(err)

	err = json.Unmarshal(bytes, &c)
	check.Error(err)
	err = c.valueCheck()
	check.Error(err)

	Threshold = c.Cluster.LoadBalancer.ThresholdOut
}

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

func (c *Config) ContainID(i int) bool {
	for _, v := range c.StartMachineIDs {
		if i == v {
			return true
		}
	}
	return false
}
