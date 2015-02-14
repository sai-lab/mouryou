package apache

import (
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var Timeout = 3

func Scoreboard(ipAddress string) string {
	var board string

	url := "http://" + ipAddress + "/server-status?auto"
	request, _ := http.NewRequest("GET", url, nil)

	client := &http.Client{Timeout: time.Duration(Timeout) * time.Second}
	response, error := client.Do(request)

	if response == nil {
		return "NIL"
	} else if error != nil {
		return "ERROR"
	}

	body, _ := ioutil.ReadAll(response.Body)

	for _, line := range strings.Split(string(body), "\n") {
		if strings.Contains(line, "Scoreboard") {
			board = line[12:]
			break
		}
	}

	defer response.Body.Close()
	return board
}

func OperatingRatio(board string) float64 {
	if board == "NIL" {
		return 0.0
	} else if board == "ERROR" {
		return 1.0
	}

	all := len(strings.Split(board, ""))
	idles := strings.Count(board, "_") + strings.Count(board, ".")

	return float64((all - idles)) / float64(all)
}
