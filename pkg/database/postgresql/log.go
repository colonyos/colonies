package postgresql

import (
	"database/sql"
	"time"

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

func (db *PQDatabase) parseLogs(rows *sql.Rows) (string, error) {
	logStr := ""

	for rows.Next() {
		var msg string
		if err := rows.Scan(&msg); err != nil {
			return logStr, err
		}
		logStr += msg
	}

	return logStr, nil
}

func (db *PQDatabase) GetLogsByProcessID(processID string, limit int) (string, error) {
	sqlStatement := `SELECT MSG FROM ` + db.dbPrefix + `LOGS WHERE PROCESS_ID=$1 ORDER BY TS ASC LIMIT $2`
	rows, err := db.postgresql.Query(sqlStatement, processID, limit)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	logStr, err := db.parseLogs(rows)
	if err != nil {
		return "", err
	}

	return logStr, nil
}

func (db *PQDatabase) DeleteLogs(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `LOGS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	return nil
}
