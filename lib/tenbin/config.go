package tenbin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func LoadConfig() Hypervisor {
	var hypervisor Hypervisor
	contents, _ := ioutil.ReadFile(os.Getenv("HOME") + "/.tenbin.json")

	if contents == nil {
		fmt.Println("Cannot open ~/.tenbin.json")
		os.Exit(1)
	}

	json.Unmarshal(contents, &hypervisor)
	return hypervisor
}
