package mouryou

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func LoadConfig() cluster {
	var c cluster
	contents, _ := ioutil.ReadFile(os.Getenv("HOME") + "/.mouryou.json")

	if contents == nil {
		fmt.Println("Cannot open ~/.mouryou.json")
		os.Exit(1)
	}

	json.Unmarshal(contents, &c)
	http.DefaultClient.Timeout = time.Duration(c.Timeout) * time.Second

	return c
}
