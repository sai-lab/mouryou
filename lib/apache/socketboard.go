package apache

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func Socketboard(host string) ([]byte, error) {
	var board []byte

	url := "http://" + host + ":8081"

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

func SocketNum(board []byte) float64 {
	var status SocketStatus

	err := json.Unmarshal(board, &status)
	if err != nil {
		return 0
	}

	return status.Socket

}
