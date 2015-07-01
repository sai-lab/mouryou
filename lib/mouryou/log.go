package mouryou

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func CreateLog() *os.File {
	now := time.Now().Format("20060102150405")
	file, err := os.OpenFile("./log/"+now+".csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	checkError(err)

	return file
}

func logging(xs []float64) {
	arr := atoa(xs)
	fmt.Println(strings.Join(arr, "  "))
	log.Println("," + strings.Join(arr, ","))
}
