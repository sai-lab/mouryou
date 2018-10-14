package databases

import (
	"encoding/json"
	"time"

	"strings"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/sai-lab/mouryou/lib/apache"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

func Connect(config *models.Config) error {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.InfluxDBAddr + ":" + config.InfluxDBPort,
		Username: config.InfluxDBUser,
		Password: config.InfluxDBPasswd,
	})
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}

	config.InfluxDBConnection = c
	return err
}

func WriteValues(clnt client.Client, config *models.Config, tag []string, field []string) {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.InfluxDBServerDB,
		Precision: "ms",
	})
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}

	var tags map[string]string
	for _, t := range tag {
		parts := strings.Split(t, ":")
		tags[parts[0]] = parts[1]
	}

	var fields map[string]interface{}
	for _, f := range field {
		parts := strings.Split(f, ":")
		fields[parts[0]] = parts[1]
	}

	pt, err := client.NewPoint(
		"server_status",
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
	bp.AddPoint(pt)

	if err := clnt.Write(bp); err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
}

func WritePoints(clnt client.Client, config *models.Config, status apache.ServerStatus) {
	var throughput float64
	var beforeApacheTime time.Time
	var beforeTotalRequest float64
	var nowApacheTime time.Time
	var nowTotalRequest float64

	nowApacheTime, err := time.Parse(time.RFC3339Nano, status.ApacheAcquisitionTime)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
	nowTotalRequest = status.ApacheStat

	query := "SELECT apache_acquisition_time, total_request FROM " + config.InfluxDBServerDB + " WHERE host = '" + status.HostName + "' AND total_request > 0 LIMIT 1"
	res, err := QueryDB(config.InfluxDBConnection, query, config.InfluxDBServerDB)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}

	if res[0].Series != nil {
		for _, row := range res[0].Series[0].Values {
			beforeApacheTime, err = time.Parse(time.RFC3339Nano, row[1].(string))
			if err != nil {
				place := logger.Place()
				logger.Error(place, err)
			}
			beforeTotalRequest, err = row[2].(json.Number).Float64()
			if err != nil {
				place := logger.Place()
				logger.Error(place, err)
			}
		}
		throughput = (nowTotalRequest - beforeTotalRequest) / (nowApacheTime.Sub(beforeApacheTime).Seconds())
	} else {
		throughput = nowTotalRequest
	}

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  config.InfluxDBServerDB,
		Precision: "ms",
	})
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}

	tags := map[string]string{
		"host":    status.HostName,
		"host_id": status.HostID,
		"vendor":  "azure",
	}

	fields := map[string]interface{}{
		"throughput":              throughput,
		"operating_ratio":         status.ApacheStat,
		"total_request":           status.ApacheLog,
		"request_per_second":      status.ReqPerSec,
		"apache_acquisition_time": status.ApacheAcquisitionTime,
		"cpu_used_percent":        status.CpuUsedPercent,
		"cpu_acquisition_time":    status.CpuUsedPercent,
		"dstat":                   status.DstatLog,
		"dstat_acquisition_time":  status.DstatAcquisitionTime,
		"disk_io":                 status.DiskIO,
		"disk_acquisition_time":   status.DiskAcquisitionTime,
		"memory_stat":             status.MemStat,
		"memory_acquisition_time": status.MemoryAcquisitionTime,
	}

	t, err := time.Parse(time.RFC3339Nano, status.Time)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
	pt, err := client.NewPoint(
		"server_status",
		tags,
		fields,
		t,
	)
	if err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
	bp.AddPoint(pt)

	if err := clnt.Write(bp); err != nil {
		place := logger.Place()
		logger.Error(place, err)
	}
}

// QueryDB convenience function to query the database
func QueryDB(clnt client.Client, cmd string, db string) ([]client.Result, error) {
	var res []client.Result
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
