package postgresql

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddRuntime(runtime *core.Runtime) error {
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `RUNTIMES (RUNTIME_ID, RUNTIME_TYPE, NAME, COLONY_ID, CPU, CORES, MEM, GPU, GPUS, STATE, COMMISSIONTIME, LASTHEARDFROM) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := db.postgresql.Exec(sqlStatement, runtime.ID, runtime.RuntimeType, runtime.Name, runtime.ColonyID, runtime.CPU, runtime.Cores, runtime.Mem, runtime.GPU, runtime.GPUs, 0, time.Now(), runtime.LastHeardFromTime)
	if err != nil {
		if strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint") {
			return errors.New("Runtime name has to be unique")
		}
		return err
	}

	return nil
}

func (db *PQDatabase) parseRuntimes(rows *sql.Rows) ([]*core.Runtime, error) {
	var runtimes []*core.Runtime

	for rows.Next() {
		var id string
		var runtimeType string
		var name string
		var colonyID string
		var cpu string
		var cores int
		var mem int
		var gpu string
		var gpus int
		var state int
		var commissionTime time.Time
		var lastHeardFromTime time.Time
		if err := rows.Scan(&id, &runtimeType, &name, &colonyID, &cpu, &cores, &mem, &gpu, &gpus, &state, &commissionTime, &lastHeardFromTime); err != nil {
			return nil, err
		}

		runtime := core.CreateRuntimeFromDB(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, state, commissionTime, lastHeardFromTime)
		runtimes = append(runtimes, runtime)
	}

	return runtimes, nil
}

func (db *PQDatabase) GetRuntimes() ([]*core.Runtime, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RUNTIMES`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseRuntimes(rows)
}

func (db *PQDatabase) GetRuntimeByID(runtimeID string) (*core.Runtime, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RUNTIMES WHERE RUNTIME_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, runtimeID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	runtimes, err := db.parseRuntimes(rows)
	if err != nil {
		return nil, err
	}

	if len(runtimes) > 1 {
		return nil, errors.New("Expected one runtime, runtime id should be unique")
	}

	if len(runtimes) == 0 {
		return nil, nil
	}

	return runtimes[0], nil
}

func (db *PQDatabase) GetRuntimesByColonyID(colonyID string) ([]*core.Runtime, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `RUNTIMES WHERE COLONY_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	runtimes, err := db.parseRuntimes(rows)
	if err != nil {
		return nil, err
	}

	return runtimes, nil
}

func (db *PQDatabase) ApproveRuntime(runtime *core.Runtime) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `RUNTIMES SET STATE=1 WHERE RUNTIME_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, runtime.ID)
	if err != nil {
		return err
	}

	runtime.Approve()

	return nil
}

func (db *PQDatabase) RejectRuntime(runtime *core.Runtime) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `RUNTIMES SET STATE=2 WHERE RUNTIME_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, runtime.ID)
	if err != nil {
		return err
	}

	runtime.Reject()

	return nil
}

func (db *PQDatabase) MarkAlive(runtime *core.Runtime) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `RUNTIMES SET LASTHEARDFROM=$1 WHERE RUNTIME_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, time.Now(), runtime.ID)
	if err != nil {
		return err
	}

	runtime.Reject()

	return nil
}

func (db *PQDatabase) DeleteRuntimeByID(runtimeID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RUNTIMES WHERE RUNTIME_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, runtimeID)
	if err != nil {
		return err
	}

	// Move back the runtime currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_RUNTIME_ID=$3, STATE=$4 WHERE ASSIGNED_RUNTIME_ID=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, runtimeID, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteRuntimesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `RUNTIMES WHERE COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
	if err != nil {
		return err
	}

	// Move back the runtime currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_RUNTIME_ID=$3, STATE=$4 WHERE TARGET_COLONY_ID=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, colonyID, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) CountRuntimes() (int, error) {
	runtimes, err := db.GetRuntimes()
	if err != nil {
		return -1, err
	}

	return len(runtimes), nil
}

func (db *PQDatabase) CountRuntimesByColonyID(colonyID string) (int, error) {
	runtimes, err := db.GetRuntimesByColonyID(colonyID)
	if err != nil {
		return -1, err
	}

	return len(runtimes), nil
}
