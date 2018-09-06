package apache

import (
	"github.com/shirou/gopsutil/mem"
)

// ServerStatus は各Webサーバの負荷状況を格納します。
type ServerStatus struct {
	// Host
	HostName             string `json:"hostname"`
	HostID               string `json:"hostid"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	// Memory
	MemStat               mem.VirtualMemoryStat `json:"memStat"`
	MemoryAcquisitionTime string                `json:"memoryAcquisitionTime"`
	// DiskIO
	DiskIO              []DiskStatus `json:"diskIO"`
	DiskAcquisitionTime string       `json:"diskAcquisitionTime"`
	// CPU
	CpuUsedPercent     []float64 `json:"cpuUsedPercent"`
	CpuAcquisitionTime string    `json:"cpuAcquisitionTime"`
	// Apache
	ApacheStat            float64 `json:"apacheStat"`
	ApacheLog             int     `json:"apacheLog"`
	ReqPerSec             float64 `json:"reqPerSec"`
	ApacheAcquisitionTime string  `json:"apacheAcquisitionTime"`
	// Dstat
	DstatLog             string `json:"dstatLog"`
	DstatAcquisitionTime string `json:"dstatAcquisitionTime"`
	// Time
	Time string `json:"time"`
	// Error
	//ErrorInfo []error `json:"errorInfo"`
	// Other Error
	Other string `json:"other"`
	// ID
	Id int `json:"-"`
}

// DiskStatus はディスクの付加情報を格納します。
type DiskStatus struct {
	Name       string `json:"name"`
	IoTime     uint64 `json:"ioTime"`
	WeightedIO uint64 `json:"weightedIO"`
}
