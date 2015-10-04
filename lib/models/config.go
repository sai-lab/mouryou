package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sai-lab/mouryou/lib/check"
)

type configStruct struct {
	Cluster ClusterStruct `json:"cluster"`
}

func LoadConfig() *ClusterStruct {
	var config configStruct

	bytes, err := ioutil.ReadFile(os.Getenv("HOME") + "/.mouryou.json")
	check.Error(err)

	err = json.Unmarshal(bytes, &config)
	check.Error(err)

	return &config.Cluster
}
