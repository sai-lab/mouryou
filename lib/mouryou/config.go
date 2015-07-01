package mouryou

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type configStruct struct {
	Cluster ClusterStruct `json:"cluster"`
}

var wait time.Duration

func LoadConfig() *ClusterStruct {
	var config configStruct

	bytes, err := ioutil.ReadFile(os.Getenv("HOME") + "/.mouryou.json")
	checkError(err)

	err = json.Unmarshal(bytes, &config)
	checkError(err)

	return &config.Cluster
}
