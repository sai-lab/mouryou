package models

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/sai-lab/mouryou/lib/check"
)

type ConfigStruct struct {
	Cluster   ClusterStruct   `json:"cluster"`
	Timeout   time.Duration   `json:"timeout"`
	Sleep     time.Duration   `json:"sleep"`
	Wait      time.Duration   `json:"wait"`
	Margin    float64         `json:"margin"`
	WebSocket WebSocketStruct `json:"web_socket"`
}

func LoadConfig(path string) *ConfigStruct {
	var config ConfigStruct

	bytes, err := ioutil.ReadFile(path)
	check.Error(err)

	err = json.Unmarshal(bytes, &config)
	check.Error(err)

	threshold = config.Cluster.LoadBalancer.ThresholdOut

	return &config
}
