package apache

type SocketStatus struct {
	// Host
	HostName             string `json:"hostname"`
	HostID               string `json:"hostid"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	// Socket
	Socket float64 `json:"socket"`
	// Time
	Time string `json:"time"`
	// Error
	ErrorInfo []error `json:"errorInfo"`
	// Other Error
	Other string `json:"other"`
	// ID
	Id int `json:"-"`
}
