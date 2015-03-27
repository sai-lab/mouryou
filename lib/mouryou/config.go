package mouryou

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func LoadConfig() cluster {
	contents, err := ioutil.ReadFile(os.Getenv("HOME") + "/.mouryou.json")
	checkError(err)

	var c cluster
	json.Unmarshal(contents, &c)
	c.init()

	http.DefaultClient.Timeout = time.Duration(c.Timeout) * time.Second

	return c
}
