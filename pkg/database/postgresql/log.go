package postgresql

import (
	"database/sql"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddLog(processID string, colonyID string, executorID string, msg string) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `LOGS (PROCESS_ID, COLONY_ID, EXECUTOR_ID, TS, MSG) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.postgresql.Exec(sqlStatement, processID, colonyID, executorID, time.Now(), msg)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseLogs(rows *sql.Rows) ([]core.Log, error) {
	var logs []core.Log

	for rows.Next() {
		var processID string
		var colonyID string
		var executorID string
		var ts time.Time
		var msg string
		if err := rows.Scan(&processID, &colonyID, &executorID, &ts, &msg); err != nil {
			return nil, err
		}
		log := core.Log{ProcessID: processID, ColonyID: colonyID, ExecutorID: executorID, Timestamp: ts, Message: msg}
		logs = append(logs, log)
	}

	return logs, nil
}

func (db *PQDatabase) GetLogsByProcessID(processID string, limit int) ([]core.Log, error) {
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

func (db *PQDatabase) DeleteLogs(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `LOGS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
