package engine

import (
	"sync"
)

const LING_SIZE = 10

var (
	working                int = 1
	booting                int = 0
	shuting                int = 0
	waiting                int = 0
	totalWeight            int = 0
	futureTotalWeight      int = 0
	workMutex              sync.RWMutex
	bootMutex              sync.RWMutex
	shutMutex              sync.RWMutex
	waitMutex              sync.RWMutex
	totalWeightMutex       sync.RWMutex
	futureTotalWeightMutex sync.RWMutex
)
