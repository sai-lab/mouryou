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
	LoadCh            = make(chan float64, 1)
	dataCh            = make(chan []DataStruct, 1)
	beforeTime        = map[string]int{}
	beforeTotalAccess = map[string]int{}
)
