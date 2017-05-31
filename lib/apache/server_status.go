package apache

type ServerStat struct {
	// Host
	HostName             string `json:"hostname"`
	HostID               string `json:"hostid"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	// Memory
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"usedPercent"`
	// DiskIO
	DiskIO []DiskStat `json:"diskIO"`
	// Cpu
	CpuUsedPercent []float64 `json:"cpuUsedPercent"`
	// Apache
	ApacheStat float64 `json:"apacheStat"`
	// Time
	Time string `json:"time"`
	// Error
	ErrorInfo []error `json:"errorInfo"`
	// Other Error
	Other string `json:"other"`
}

type DiskStat struct {
	Name       string `json:"name"`
	IoTime     uint64 `json:"ioTime"`
	WeightedIO uint64 `json:"weightedIO"`
}
