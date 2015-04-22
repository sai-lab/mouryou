package apache

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

func Scoreboard(ipAddress string) (string, error) {
	var board string

	url := "http://" + ipAddress + "/server-status?auto"
	request, _ := http.NewRequest("GET", url, nil)
	response, err := http.DefaultClient.Do(request)

	if response == nil {
		return board, errors.New("apache: no response")
	} else if err != nil {
		return board, errors.New("apache: request timeout")
	}

	body, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()

	lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")
	board = lines[len(lines)-1][12:]
	return board, nil
}

func OperatingRatio(board string) float64 {
	all := len(strings.Split(board, ""))
	idles := strings.Count(board, "_") + strings.Count(board, ".")

	return float64((all - idles)) / float64(all)
}
