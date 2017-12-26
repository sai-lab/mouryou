package monitor

type StatusStruct struct {
	Name   string
	Weight int
	Info   string
}

type DataStruct struct {
	Name       string
	Operating  float64
	Throughput int
	Cpu        float64
}

type PowerStruct struct {
	Name string
	Info string
}

var (
	StatusCh          = make(chan StatusStruct, 1)
	PowerCh           = make(chan PowerStruct, 1)
	LoadCh            = make(chan float64, 1)
	DataCh            = make(chan []DataStruct, 1)
	States            []StatusStruct
	beforeTime        = map[string]int{}
	beforeTotalAccess = map[string]int{}
)
