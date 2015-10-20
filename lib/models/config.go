package models

import (
	"encoding/json"
	"io/ioutil"

	"github.com/sai-lab/mouryou/lib/check"
)

type configStruct struct {
	Cluster ClusterStruct `json:"cluster"`
}

func LoadConfig(path string) *ClusterStruct {
	var config configStruct

	bytes, err := ioutil.ReadFile(path)
	check.Error(err)

	err = json.Unmarshal(bytes, &config)
	check.Error(err)

	return &config.Cluster
}
