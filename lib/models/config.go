package models

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/sai-lab/mouryou/lib/check"
)

type Config struct {
	Timeout         time.Duration   `json:"timeout"`
	Sleep           time.Duration   `json:"sleep"`
	Wait            time.Duration   `json:"wait"`
	Margin          float64         `json:"margin"`
	Algorithm       string          `json:"algorithm"`
	UseHetero       bool            `json:"use_hetero"`
	AdjustServerNum bool            `json:"adjust_server_num"`
	WebSocket       WebSocketStruct `json:"web_socket"`
	Cluster         Cluster  `json:"cluster"`
}

//LoadConfig
func (c *Config) LoadSetting(path string) {

	bytes, err := ioutil.ReadFile(path)
	check.Error(err)

	err = json.Unmarshal(bytes, &c)
	check.Error(err)

	Threshold = c.Cluster.LoadBalancer.ThresholdOut

}
