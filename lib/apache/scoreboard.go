package apache

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func Scoreboard(ipAddress string) string {
	url := "http://" + ipAddress + "/server-status?auto"
	request, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	response, _ := client.Do(request)

	if response == nil {
		return ""
	} else {
		body, _ := ioutil.ReadAll(response.Body)
		line := strings.Split(string(body), "\n")[9]
		defer response.Body.Close()
		return line[12:]
	}
}

func OperatingRatio(board string) float64 {
	all := len(strings.Split(board, ""))
	idles := strings.Count(board, "_") + strings.Count(board, ".")

	return float64((all - idles)) / float64(all)
}
