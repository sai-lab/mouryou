package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
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

func PWArrays(developLogLevel int, arrs [11][]string) {
	for i := 0; i < 11; i++ {
		if i == 3 {
			continue
		}
		if i == 8 || i == 7 {
			log.Println("," + strings.Join(arrs[i], ","))
			continue
		}
		if developLogLevel >= 3 {
			// サーバのパラメータを全て標準出力に出力
			fmt.Println(strings.Join(arrs[i], " "))
		}

		log.Println("," + strings.Join(arrs[i], ","))
	}
}

func PrintPlace(str string) {
	var i int = 0
	var path string

	_, file, line, _ := runtime.Caller(1)
	files := strings.Split(file, "/")

	for i = 0; i < len(files); i++ {
		if files[i] == "mouryou" {
			break
		}
	}

	if i+1 == len(files) {
		path = file
	} else {
		path = strings.Join(files[i+1:], "/")
	}

	fmt.Println(path + " Line " + fmt.Sprint(line) + " " + str)
}

func Send(connection *websocket.Conn, err error, data interface{}) {
	if err != nil {
		return
	}

	// var message string

	// switch data.(type) {
	// case string:
	// 	message = data.(string)
	// case []string:
	// 	message = "Loads: " + strings.Join(data.([]string), ",")
	// }

	// websocket.Message.Send(connection, message)
}
