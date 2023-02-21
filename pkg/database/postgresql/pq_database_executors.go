package postgresql

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddExecutor(executor *core.Executor) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `EXECUTORS (EXECUTOR_ID, EXECUTOR_TYPE, NAME, COLONY_ID, STATE, COMMISSIONTIME, LASTHEARDFROM, LONG, LAT) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := db.postgresql.Exec(sqlStatement, executor.ID, executor.Type, executor.Name, executor.ColonyID, 0, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat)
	if err != nil {
		if strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint") {
			return errors.New("Executor name must be unique")
		}
		return err
	}

	return nil
}

func (db *PQDatabase) parseExecutors(rows *sql.Rows) ([]*core.Executor, error) {
	var executors []*core.Executor

	for rows.Next() {
		var id string
		var executorType string
		var name string
		var colonyID string
		var state int
		var commissionTime time.Time
		var lastHeardFromTime time.Time
		var long float64
		var lat float64
		if err := rows.Scan(&id, &executorType, &name, &colonyID, &state, &commissionTime, &lastHeardFromTime, &long, &lat); err != nil {
			return nil, err
		}

		executor := core.CreateExecutorFromDB(id, executorType, name, colonyID, state, commissionTime, lastHeardFromTime)
		executor.Location.Long = long
		executor.Location.Lat = lat
		executors = append(executors, executor)
	}

	return executors, nil
}

func (db *PQDatabase) GetExecutors() ([]*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseExecutors(rows)
}

func (db *PQDatabase) GetExecutorByID(executorID string) (*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE EXECUTOR_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, executorID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	executors, err := db.parseExecutors(rows)
	if err != nil {
		return nil, err
	}

	if len(executors) > 1 {
		return nil, errors.New("Expected one executor, executor id should be unique")
	}

	if len(executors) == 0 {
		return nil, nil
	}

	return executors[0], nil
}

func (db *PQDatabase) GetExecutorsByColonyID(colonyID string) ([]*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	executors, err := db.parseExecutors(rows)
	if err != nil {
		return nil, err
	}

	return executors, nil
}

func (db *PQDatabase) ApproveExecutor(executor *core.Executor) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET STATE=1 WHERE EXECUTOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, executor.ID)
	if err != nil {
		return err
	}

	executor.Approve()

	return nil
}

func (db *PQDatabase) RejectExecutor(executor *core.Executor) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET STATE=2 WHERE EXECUTOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, executor.ID)
	if err != nil {
		return err
	}

	executor.Reject()

	return nil
}

func (db *PQDatabase) MarkAlive(executor *core.Executor) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET LASTHEARDFROM=$1 WHERE EXECUTOR_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, time.Now(), executor.ID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteExecutorByID(executorID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `EXECUTORS WHERE EXECUTOR_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, executorID)
	if err != nil {
		return err
	}

	// Move back the executor currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_EXECUTOR_ID=$3, STATE=$4 WHERE ASSIGNED_EXECUTOR_ID=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, executorID, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteExecutorsByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	// Move back the executor currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_EXECUTOR_ID=$3, STATE=$4 WHERE TARGET_COLONY_ID=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, colonyID, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountExecutors() (int, error) {
	executors, err := db.GetExecutors()
	if err != nil {
		return -1, err
	}

	return len(executors), nil
}

func (db *PQDatabase) CountExecutorsByColonyID(colonyID string) (int, error) {
	executors, err := db.GetExecutorsByColonyID(colonyID)
	if err != nil {
		return -1, err
	}

	return len(executors), nil
}
