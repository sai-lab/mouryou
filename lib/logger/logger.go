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

func Send(ws *websocket.Conn, arr []string) {
	websocket.Message.Send(ws, "Loads: "+strings.Join(arr, ","))
}
