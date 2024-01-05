package postgresql

import (
	"database/sql"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `LOGS (PROCESS_ID, COLONY_NAME, EXECUTOR_NAME, TS, MSG, ADDED) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.postgresql.Exec(sqlStatement, processID, colonyName, executorName, timestamp, msg, time.Now())
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) addHistoricalLog(processID string, colonyName string, executorName string, timestamp int64, msg string, t time.Time) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `LOGS (PROCESS_ID, COLONY_NAME, EXECUTOR_NAME, TS, MSG, ADDED) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.postgresql.Exec(sqlStatement, processID, colonyName, executorName, timestamp, msg, t)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseLogs(rows *sql.Rows) ([]*core.Log, error) {
	var logs []*core.Log

	for rows.Next() {
		var processID string
		var colonyName string
		var executorName string
		var ts int64
		var msg string
		var added time.Time
		if err := rows.Scan(&processID, &colonyName, &executorName, &ts, &msg, &added); err != nil {
			return nil, err
		}
		log := &core.Log{ProcessID: processID, ColonyName: colonyName, ExecutorName: executorName, Timestamp: ts, Message: msg}
		logs = append(logs, log)
	}

	return logs, nil
}

func (db *PQDatabase) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOGS WHERE PROCESS_ID=$1 ORDER BY TS ASC LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, processID, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	logs, err := db.parseLogs(rows)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (db *PQDatabase) GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOGS WHERE EXECUTOR_NAME=$1 ORDER BY TS ASC LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, executorName, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	logs, err := db.parseLogs(rows)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (db *PQDatabase) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOGS WHERE PROCESS_ID=$1 AND TS>$2 ORDER BY TS ASC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, processID, since, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	logs, err := db.parseLogs(rows)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (db *PQDatabase) GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `LOGS WHERE EXECUTOR_NAME=$1 AND TS>$2 ORDER BY TS ASC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, executorName, since, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	logs, err := db.parseLogs(rows)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (db *PQDatabase) RemoveLogsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `LOGS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountLogs(colonyName string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `LOGS WHERE COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
	if err != nil {
		return -1, err
	}

	defer rows.Close()

	rows.Next()
	var count int
	err = rows.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (db *PQDatabase) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) {
	sqlStatement := `SELECT *
                     FROM ` + db.dbPrefix + `LOGS
                     WHERE MSG LIKE '%' || $1 || '%' AND COLONY_NAME = $2 
                     AND ADDED > NOW() - make_interval(days => $3) LIMIT $4`

	rows, err := db.postgresql.Query(sqlStatement, text, colonyName, days, count)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []*core.Log
	for rows.Next() {
		var processID string
		var colonyName string
		var executorName string
		var timestamp int64
		var message string
		var added time.Time
		if err := rows.Scan(&processID, &colonyName, &executorName, &timestamp, &message, &added); err != nil {
			return nil, err
		}
		results = append(results, &core.Log{ProcessID: processID, ColonyName: colonyName, ExecutorName: executorName, Message: message, Timestamp: timestamp})
	}

	return results, nil
}
