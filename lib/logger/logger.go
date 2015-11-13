package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/sai-lab/mouryou/lib/check"
)

func Create() *os.File {
	now := time.Now().Format("20060102150405")
	file, err := os.OpenFile("./log/"+now+".csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	check.Error(err)

	return file
}

func Print(arr []string) {
	fmt.Println(strings.Join(arr, "  "))
}

func Write(arr []string) {
	log.Println("," + strings.Join(arr, ","))
}

func Send(connection *websocket.Conn, err error, data interface{}) {
	if err != nil {
		return
	}

	var message string

	switch data.(type) {
	case string:
		message = data.(string)
	case []string:
		message = "Loads: " + strings.Join(data.([]string), ",")
	}

	websocket.Message.Send(connection, message)
}
