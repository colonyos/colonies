package postgresql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *PQDatabase) createMetricsTable() error {
	sqlStatement := `CREATE TABLE IF NOT EXISTS ` + db.dbPrefix + `METRICS (METRIC_ID TEXT PRIMARY KEY NOT NULL, COLONY_NAME TEXT NOT NULL, EXECUTOR_NAME TEXT NOT NULL, KEY TEXT NOT NULL, METRIC_TYPE INTEGER NOT NULL, VALUE DOUBLE PRECISION NOT NULL, PERIOD INTEGER NOT NULL DEFAULT 0, PERIOD_START TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01T00:00:00Z')`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}
	return nil
}

func (db *PQDatabase) dropMetricsTable() error {
	sqlStatement := `DROP TABLE IF EXISTS ` + db.dbPrefix + `METRICS`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}
	return nil
}

func (db *PQDatabase) parseMetrics(rows *sql.Rows) ([]core.Metric, error) {
	var metrics []core.Metric
	for rows.Next() {
		var metricID string
		var colonyName string
		var executorName string
		var key string
		var metricType int
		var value float64
		var period int
		var periodStart time.Time
		if err := rows.Scan(&metricID, &colonyName, &executorName, &key, &metricType, &value, &period, &periodStart); err != nil {
			return nil, err
		}
		metric := core.Metric{
			ColonyName:   colonyName,
			ExecutorName: executorName,
			Key:          key,
			MetricType:   metricType,
			Value:        value,
			Period:       period,
			PeriodStart:  periodStart,
		}
		metric.GenerateID()
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (db *PQDatabase) SetMetric(metric core.Metric) error {
	metric.GenerateID()
	sqlStatement := `INSERT INTO ` + db.dbPrefix + `METRICS (METRIC_ID, COLONY_NAME, EXECUTOR_NAME, KEY, METRIC_TYPE, VALUE, PERIOD, PERIOD_START) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (METRIC_ID) DO UPDATE SET VALUE=$6, METRIC_TYPE=$5`
	_, err := db.postgresql.Exec(sqlStatement, metric.ID, metric.ColonyName, metric.ExecutorName, metric.Key, metric.MetricType, metric.Value, metric.Period, metric.PeriodStart)
	return err
}

func (db *PQDatabase) GetMetric(colonyName string, executorName string, key string, period int, periodStart time.Time) (core.Metric, error) {
	m := core.Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		Period:       period,
		PeriodStart:  periodStart,
	}
	m.GenerateID()
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `METRICS WHERE METRIC_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, m.ID)
	if err != nil {
		return core.Metric{}, err
	}
	defer rows.Close()

	metrics, err := db.parseMetrics(rows)
	if err != nil {
		return core.Metric{}, err
	}
	if len(metrics) == 0 {
		return core.Metric{}, errors.New("Metric does not exist")
	}
	return metrics[0], nil
}

func (db *PQDatabase) GetMetricsByExecutorName(colonyName string, executorName string) ([]core.Metric, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `METRICS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2 AND PERIOD=0`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, executorName)
	if err != nil {
		return []core.Metric{}, err
	}
	defer rows.Close()

	metrics, err := db.parseMetrics(rows)
	if err != nil {
		return []core.Metric{}, err
	}
	if metrics == nil {
		metrics = make([]core.Metric, 0)
	}
	return metrics, nil
}

func (db *PQDatabase) GetAllMetricsByExecutorName(colonyName string, executorName string) ([]core.Metric, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `METRICS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, executorName)
	if err != nil {
		return []core.Metric{}, err
	}
	defer rows.Close()

	metrics, err := db.parseMetrics(rows)
	if err != nil {
		return []core.Metric{}, err
	}
	if metrics == nil {
		metrics = make([]core.Metric, 0)
	}
	return metrics, nil
}

func (db *PQDatabase) GetMetricsByColonyName(colonyName string) ([]core.Metric, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `METRICS WHERE COLONY_NAME=$1 AND PERIOD=0`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return []core.Metric{}, err
	}
	defer rows.Close()

	metrics, err := db.parseMetrics(rows)
	if err != nil {
		return []core.Metric{}, err
	}
	if metrics == nil {
		metrics = make([]core.Metric, 0)
	}
	return metrics, nil
}

func (db *PQDatabase) GetMetricHistory(colonyName string, executorName string, key string, period int, from time.Time, to time.Time) ([]core.Metric, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `METRICS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2 AND KEY=$3 AND PERIOD=$4 AND PERIOD_START >= $5 AND PERIOD_START <= $6 ORDER BY PERIOD_START ASC`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, executorName, key, period, from, to)
	if err != nil {
		return []core.Metric{}, err
	}
	defer rows.Close()

	metrics, err := db.parseMetrics(rows)
	if err != nil {
		return []core.Metric{}, err
	}
	if metrics == nil {
		metrics = make([]core.Metric, 0)
	}
	return metrics, nil
}

func (db *PQDatabase) IncrementMetric(colonyName string, executorName string, key string, period int, periodStart time.Time, delta float64) error {
	m := core.Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		MetricType:   core.COUNTER,
		Period:       period,
		PeriodStart:  periodStart,
	}
	m.GenerateID()
	sqlStatement := `INSERT INTO ` + db.dbPrefix + `METRICS (METRIC_ID, COLONY_NAME, EXECUTOR_NAME, KEY, METRIC_TYPE, VALUE, PERIOD, PERIOD_START) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (METRIC_ID) DO UPDATE SET VALUE=` + db.dbPrefix + `METRICS.VALUE+$6`
	_, err := db.postgresql.Exec(sqlStatement, m.ID, colonyName, executorName, key, core.COUNTER, delta, period, periodStart)
	return err
}

func (db *PQDatabase) RemoveMetric(colonyName string, executorName string, key string, period int, periodStart time.Time) error {
	m := core.Metric{
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Key:          key,
		Period:       period,
		PeriodStart:  periodStart,
	}
	m.GenerateID()
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `METRICS WHERE METRIC_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, m.ID)
	return err
}

func (db *PQDatabase) RemoveAllMetricsByExecutorName(colonyName string, executorName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `METRICS WHERE COLONY_NAME=$1 AND EXECUTOR_NAME=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, executorName)
	return err
}

func (db *PQDatabase) RemoveAllMetricsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `METRICS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	return err
}

func (db *PQDatabase) RemoveAllMetrics() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `METRICS`
	_, err := db.postgresql.Exec(sqlStatement)
	return err
}
