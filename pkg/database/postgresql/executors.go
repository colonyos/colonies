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
	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `EXECUTORS (EXECUTOR_ID, EXECUTOR_TYPE, NAME, COLONY_ID, STATE, REQUIRE_FUNC_REG, COMMISSIONTIME, LASTHEARDFROM, LONG, LAT, LOCDESC, HWMODEL, HWCPU, HWMEM, HWSTORAGE, HWGPUNAME, HWGPUCOUNT, SWNAME, SWTYPE, SWVERSION) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)`
	_, err := db.postgresql.Exec(sqlStatement, executor.ID, executor.Type, executor.Name, executor.ColonyID, 0, executor.RequireFuncReg, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, executor.Capabilities.Hardware.Model, executor.Capabilities.Hardware.CPU, executor.Capabilities.Hardware.Memory, executor.Capabilities.Hardware.Storage, executor.Capabilities.Hardware.GPU.Name, executor.Capabilities.Hardware.GPU.Count, executor.Capabilities.Software.Name, executor.Capabilities.Software.Type, executor.Capabilities.Software.Version)
	if err != nil {
		if strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint") {
			return errors.New("Executor name must be unique")
		}
		return err
	}

	return nil
}

func (db *PQDatabase) AddOrReplaceExecutor(executor *core.Executor) error {
	sqlStatement := `INSERT INTO ` + db.dbPrefix + `EXECUTORS (EXECUTOR_ID, EXECUTOR_TYPE, NAME, COLONY_ID, STATE, REQUIRE_FUNC_REG, COMMISSIONTIME, LASTHEARDFROM, LONG, LAT, LOCDESC, HWMODEL, HWCPU, HWMEM, HWSTORAGE, HWGPUNAME, HWGPUCOUNT, SWNAME, SWTYPE, SWVERSION) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20) ON CONFLICT (EXECUTOR_ID) DO UPDATE SET EXECUTOR_TYPE=EXCLUDED.EXECUTOR_TYPE, NAME=EXCLUDED.NAME, COLONY_ID=EXCLUDED.COLONY_ID, STATE=EXCLUDED.STATE, REQUIRE_FUNC_REG=EXCLUDED.REQUIRE_FUNC_REG, COMMISSIONTIME=EXCLUDED.COMMISSIONTIME, LASTHEARDFROM=EXCLUDED.LASTHEARDFROM, LONG=EXCLUDED.LONG, LAT=EXCLUDED.LAT, LOCDESC=EXCLUDED.LOCDESC, HWMODEL=EXCLUDED.HWMODEL, HWCPU=EXCLUDED.HWCPU, HWMEM=EXCLUDED.HWMEM, HWSTORAGE=EXCLUDED.HWSTORAGE, HWGPUNAME=EXCLUDED.HWGPUNAME, HWGPUCOUNT=EXCLUDED.HWGPUCOUNT, SWNAME=EXCLUDED.SWNAME, SWTYPE=EXCLUDED.SWTYPE, SWVERSION=EXCLUDED.SWVERSION;`
	_, err := db.postgresql.Exec(sqlStatement, executor.ID, executor.Type, executor.Name, executor.ColonyID, 0, executor.RequireFuncReg, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, executor.Capabilities.Hardware.
		Model, executor.Capabilities.Hardware.CPU, executor.Capabilities.Hardware.Memory, executor.Capabilities.Hardware.Storage, executor.Capabilities.Hardware.GPU.Name, executor.Capabilities.Hardware.GPU.Count, executor.Capabilities.Software.Name, executor.Capabilities.Software.Type, executor.Capabilities.Software.Version)
	if err != nil {
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
		var requireRunReg bool
		var commissionTime time.Time
		var lastHeardFromTime time.Time
		var long float64
		var lat float64
		var desc string
		var hwModel string
		var hwCPU string
		var hwMem string
		var hwStorage string
		var hwGPUName string
		var hwGPUCount int
		var swName string
		var swType string
		var swVersion string
		if err := rows.Scan(&id, &executorType, &name, &colonyID, &state, &requireRunReg, &commissionTime, &lastHeardFromTime, &long, &lat, &desc, &hwModel, &hwCPU, &hwMem, &hwStorage, &hwGPUName, &hwGPUCount, &swName, &swType, &swVersion); err != nil {
			return nil, err
		}

		executor := core.CreateExecutorFromDB(id, executorType, name, colonyID, state, requireRunReg, commissionTime, lastHeardFromTime)
		location := core.Location{Long: long, Lat: lat, Description: desc}
		executor.Location = location
		gpu := core.GPU{Name: hwGPUName, Count: hwGPUCount}
		hw := core.Hardware{Model: hwModel, CPU: hwCPU, Memory: hwMem, Storage: hwStorage, GPU: gpu}
		sw := core.Software{Name: swName, Type: swType, Version: swVersion}
		capabilities := core.Capabilities{Hardware: hw, Software: sw}
		executor.Capabilities = capabilities

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

func (db *PQDatabase) GetExecutorByName(colonyID string, executorName string) (*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_ID=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, executorName)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	executors, err := db.parseExecutors(rows)
	if err != nil {
		return nil, err
	}

	if len(executors) == 0 {
		return nil, nil
	}

	return executors[0], nil
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

	err = db.DeleteFunctionsByExecutorID(executorID)
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

	err = db.DeleteFunctionsByColonyID(colonyID)
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
