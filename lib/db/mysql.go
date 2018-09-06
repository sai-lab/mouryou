package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sai-lab/mouryou/lib/logger"
	"github.com/sai-lab/mouryou/lib/models"
)

var (
	db         string
	dsn        string
	connection *sql.DB
)

func Set(c *models.Config) {
	db = c.LogDB
	dsn = c.LogDSN
}

func Connect() error {
	var err error
	connection, err = sql.Open(db, dsn)
	return err
}

func ThroughputInsert(serverName string, num int, unixTime int) error {
	stmt, err := connection.Prepare(`
      INSERT INTO throughputs (server_name, throughput, measurement_time) VALUES(?, ?, ?)
  `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(serverName, num, unixTime)
	return err
}

func LoadThroughput(serverName string) ([]int, []int) {
	var throughputs []int
	var measurementTimes []int

	rows, err := connection.Query(`
      SELECT throughput, measurement_time FROM throughputs 
      WHERE server_name = ? ORDER BY measurement_time DESC LIMIT 10`, serverName)
	if err != nil {
		logger.PrintPlace(fmt.Sprint(err))
	}
	defer rows.Close()

	for rows.Next() {
		throughput, measurementTime := 0, 0
		err = rows.Scan(&throughput, &measurementTime)
		if err != nil {
			logger.PrintPlace(fmt.Sprint(err))
		}
		throughputs = append(throughputs, throughput)
		measurementTimes = append(measurementTimes, measurementTime)
	}

	return throughputs, measurementTimes
}
