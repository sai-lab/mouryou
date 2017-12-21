package engine

import (
	"sync"
)

const LING_SIZE = 10

type StatusStruct struct {
	Name   string
	Weight int
	Info   string
}

var (
	StatusCh         = make(chan StatusStruct, 1)
	PowerCh          = make(chan PowerStruct, 1)
	States           []StatusStruct
	working          int = 1
	booting          int = 0
	shuting          int = 0
	waiting          int = 0
	totalWeight      int = 0
	workMutex        sync.RWMutex
	bootMutex        sync.RWMutex
	shutMutex        sync.RWMutex
	waitMutex        sync.RWMutex
	totalWeightMutex sync.RWMutex
)
