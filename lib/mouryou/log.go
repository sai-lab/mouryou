package mouryou

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func CreateLog() *os.File {
	now := time.Now().Format("20060102150405")
	f, err := os.OpenFile("./log/"+now+".csv", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	checkError(err)
	return f
}

func sliceToCsv(arr []float64) string {
	str := fmt.Sprintf("%+v", arr)

	str = strings.TrimLeft(str, "[")
	str = strings.TrimRight(str, "]")
	str = strings.Replace(str, " ", ",", -1)

	return str
}
