package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sai-lab/mouryou/lib/check"
	"github.com/sai-lab/mouryou/lib/convert"
)

func Create() *os.File {
	now := time.Now().Format("20060102150405")
	file, err := os.OpenFile("./log/"+now+".csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	check.Error(err)

	return file
}

func Print(xs []float64) {
	arr := convert.FloatsToStrings(xs)
	fmt.Println(strings.Join(arr, "  "))
}

func Write(xs []float64) {
	arr := convert.FloatsToStrings(xs)
	log.Println("," + strings.Join(arr, ","))
}
