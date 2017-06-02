package functions

import (
	"sync"
)

const LING_SIZE = 10

type PowerStruct struct {
	Name string
	Info string
}

type StatusStruct struct {
	Name   string
	Weight float64
	Info   string
}

type CriticalStruct struct {
	Name string
	Info string
}

var (
	loadCh           = make(chan float64, 1)
	powerCh          = make(chan PowerStruct, 1)
	statusCh         = make(chan StatusStruct, 1)
	criticalCh       = make(chan CriticalStruct, 1)
	states           []StatusStruct
	working          int     = 1
	booting          int     = 0
	shuting          int     = 0
	waiting          int     = 0
	totalWeight      float64 = 10
	workMutex        sync.RWMutex
	bootMutex        sync.RWMutex
	shutMutex        sync.RWMutex
	waitMutex        sync.RWMutex
	totalWeightMutex sync.RWMutex
)
