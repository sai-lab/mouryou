package tenbin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func LoadConfig() cluster {
	var c cluster
	contents, _ := ioutil.ReadFile(os.Getenv("HOME") + "/.tenbin.json")

	if contents == nil {
		fmt.Println("Cannot open ~/.tenbin.json")
		os.Exit(1)
	}

	json.Unmarshal(contents, &c)
	return c
}
