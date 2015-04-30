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
	f, err := os.OpenFile("./log/"+now+".csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	checkError(err)

	return f
}

func logging(ors []float64) {
	arr := formatOrs(ors)
	fmt.Println(strings.Join(arr, "  "))
	log.Println("," + strings.Join(arr, ","))
}

func formatOrs(ors []float64) []string {
	arr := make([]string, len(ors))

	for i, v := range ors {
		arr[i] = fmt.Sprintf("%.5f", v)
	}

	return arr
}
