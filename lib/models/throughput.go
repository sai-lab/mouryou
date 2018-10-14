package models

import "time"

type ThroughputWithTime struct {
	Throughput      float64
	MeasurementTime time.Time
}
