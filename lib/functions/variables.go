package functions

import "sync"

const LING_SIZE = 10

var (
	loadCh        = make(chan float64, 1)
	powerCh       = make(chan string, 1)
	working   int = 1
	booting   int = 0
	shuting   int = 0
	waiting   int = 0
	workMutex sync.RWMutex
	bootMutex sync.RWMutex
	shutMutex sync.RWMutex
	waitMutex sync.RWMutex
)
