package check

import (
	"fmt"
	"os"
)

func Error(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
