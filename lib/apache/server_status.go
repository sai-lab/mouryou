package apache

import (
	"github.com/shirou/gopsutil/mem"
)

type ServerStat struct {
	// Host
	HostName             string `json:"hostname"`
	HostID               string `json:"hostid"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	// Memory
	MemStat mem.VirtualMemoryStat `json:"memStat"`
	// DiskIO
	DiskIO []DiskStat `json:"diskIO"`
	// Cpu
	CpuUsedPercent []float64 `json:"cpuUsedPercent"`
	// Apache
	ApacheStat float64 `json:"apacheStat"`
	ApacheLog  int     `json:"apacheLog"`
	ReqPerSec  float64 `json:"reqPerSec"`
	// Dstat
	DstatLog string `json:"dstatLog"`
	// Time
	Time string `json:"time"`
	// Error
	//ErrorInfo []error `json:"errorInfo"`
	// Other Error
	Other string `json:"other"`
	// ID
	Id int `json:"-"`
}

type DiskStat struct {
	Name       string `json:"name"`
	IoTime     uint64 `json:"ioTime"`
	WeightedIO uint64 `json:"weightedIO"`
}
