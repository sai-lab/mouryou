package mouryou

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type config struct {
	Cluster cluster
	Timeout int
	Wait    int
}

var wait time.Duration

func LoadConfig() cluster {
	contents, err := ioutil.ReadFile(os.Getenv("HOME") + "/.mouryou.json")
	checkError(err)

	var c config
	json.Unmarshal(contents, &c)

	c.Cluster.init()
	http.DefaultClient.Timeout = time.Duration(c.Timeout) * time.Second
	wait = time.Duration(c.Wait)

	return c.Cluster
}
