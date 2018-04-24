package apache

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func Scoreboard(host string) ([]byte, error) {
	var board []byte

	url := "http://" + host + ":8080"

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return board, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return board, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return board, err
	}
	defer response.Body.Close()

	board = body

	return board, nil
}

func OperatingRatio(board []byte) float64 {
	var status ServerStatus
	err := json.Unmarshal(board, &status)
	if err != nil {
		return 0
	}

	return status.ApacheStat
}

// func OperatingRatio(board string) float64 {
// 	all := len(strings.Split(board, ""))
// 	idles := strings.Count(board, "_") + strings.Count(board, ".")

// 	return float64((all - idles)) / float64(all)
// }

// func Scoreboard(host string) (string, error) {
// 	var board string
// 	url := "http://" + host + "/server-status?auto"

// 	request, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return board, err
// 	}

// 	response, err := http.DefaultClient.Do(request)
// 	if err != nil {
// 		return board, err
// 	}

// 	body, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		return board, err
// 	}
// 	defer response.Body.Close()

// 	lines := strings.Split(strings.TrimRight(string(body), "\n"), "\n")
// 	board = lines[len(lines)-1][12:]

// 	return board, nil
// }
