package engine

import (
	"sync"
)

const LING_SIZE = 10

type Scale struct {
	Handle string
	Weight int
	Load   string
}

var (
	working                int = 1
	booting                int = 0
	shutting               int = 0
	waiting                int = 0
	totalWeight            int = 0
	futureTotalWeight      int = 0
	workMutex              sync.RWMutex
	bootMutex              sync.RWMutex
	shutMutex              sync.RWMutex
	waitMutex              sync.RWMutex
	totalWeightMutex       sync.RWMutex
	futureTotalWeightMutex sync.RWMutex

	scaleCh = make(chan Scale, 1)
)
