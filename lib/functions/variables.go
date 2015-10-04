package functions

import "sync"

const (
	LING_SIZE    = 10
	TIMEOUT_SEC  = 1
	SLEEP_SEC    = 30
	RATIO_MARGIN = 0.007
)

var (
	loadCh           = make(chan float64, 1)
	powerCh          = make(chan string, 1)
	working      int = 1
	operating    int = 0
	waiting      int = 0
	workMutex    sync.RWMutex
	operateMutex sync.RWMutex
	waitMutex    sync.RWMutex
)
