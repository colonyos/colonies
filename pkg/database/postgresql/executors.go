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
	if executor == nil {
		return errors.New("Executor is nil")
	}

	existingExecutor, err := db.GetExecutorByName(executor.ColonyName, executor.Name)

	if err != nil {
		return err
	}

	if existingExecutor != nil {
		return errors.New("Executor with name <" + executor.Name + "> already exists in Colony with name <" + executor.ColonyName + ">")
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `EXECUTORS (NAME, EXECUTOR_TYPE, EXECUTOR_ID, COLONY_NAME, STATE, REQUIRE_FUNC_REG, COMMISSIONTIME, LASTHEARDFROM, LONG, LAT, LOCDESC, HWMODEL, HWNODES, HWCPU, HWMEM, HWSTORAGE, HWGPUNAME, HWGPUCOUNT, HWGPUNODECOUNT, HWGPUMEM, SWNAME, SWTYPE, SWVERSION) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`
	_, err = db.postgresql.Exec(sqlStatement, executor.ColonyName+":"+executor.Name, executor.Type, executor.ID, executor.ColonyName, 0, executor.RequireFuncReg, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, executor.Capabilities.Hardware.Model, executor.Capabilities.Hardware.Nodes, executor.Capabilities.Hardware.CPU, executor.Capabilities.Hardware.Memory, executor.Capabilities.Hardware.Storage, executor.Capabilities.Hardware.GPU.Name, executor.Capabilities.Hardware.GPU.Count, executor.Capabilities.Hardware.GPU.NodeCount, executor.Capabilities.Hardware.GPU.Memory, executor.Capabilities.Software.Name, executor.Capabilities.Software.Type, executor.Capabilities.Software.Version)

	if err != nil {
		if strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint") {
			return errors.New("Executor not unique, both Name and ExecutorId must be unique within a Colony")
		}
		return err
	}

	return nil
}

func (db *PQDatabase) parseExecutors(rows *sql.Rows) ([]*core.Executor, error) {
	var executors []*core.Executor

	for rows.Next() {
		var name string
		var executorType string
		var id string
		var colonyName string
		var state int
		var requireRunReg bool
		var commissionTime time.Time
		var lastHeardFromTime time.Time
		var long float64
		var lat float64
		var desc string
		var hwModel string
		var hwNodes int
		var hwCPU string
		var hwMem string
		var hwStorage string
		var hwGPUName string
		var hwGPUCount int
		var hwGPUNodeCount int
		var hwGPUMem string
		var swName string
		var swType string
		var swVersion string
		if err := rows.Scan(&name, &executorType, &id, &colonyName, &state, &requireRunReg, &commissionTime, &lastHeardFromTime, &long, &lat, &desc, &hwModel, &hwNodes, &hwCPU, &hwMem, &hwStorage, &hwGPUName, &hwGPUCount, &hwGPUNodeCount, &hwGPUMem, &swName, &swType, &swVersion); err != nil {
			return nil, err
		}

		s := strings.Split(name, ":")
		if len(s) != 2 {
			return nil, errors.New("Failed to parse Executor name")
		}
		name = s[1]

		executor := core.CreateExecutorFromDB(id, executorType, name, colonyName, state, requireRunReg, commissionTime, lastHeardFromTime)
		location := core.Location{Long: long, Lat: lat, Description: desc}
		executor.Location = location
		gpu := core.GPU{Name: hwGPUName, Count: hwGPUCount, Memory: hwGPUMem, NodeCount: hwGPUNodeCount}
		hw := core.Hardware{Model: hwModel, CPU: hwCPU, Memory: hwMem, Storage: hwStorage, GPU: gpu, Nodes: hwNodes}
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

func (db *PQDatabase) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_NAME=$1`
	rows, err := db.postgresql.Query(sqlStatement, colonyName)
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

func (db *PQDatabase) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_NAME=$1 AND NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, colonyName+":"+executorName)
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
	executor, err := db.GetExecutorByID(executorID)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `EXECUTORS WHERE EXECUTOR_ID=$1`
	_, err = db.postgresql.Exec(sqlStatement, executorID)
	if err != nil {
		return err
	}

	// Move back the executor currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_EXECUTOR_ID=$3, STATE=$4 WHERE ASSIGNED_EXECUTOR_ID=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, executorID, core.RUNNING)
	if err != nil {
		return err
	}

	err = db.DeleteFunctionsByExecutorName(executor.ColonyName, executor.Name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteExecutorsByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_NAME=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	// Move back the executor currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_EXECUTOR_ID=$3, STATE=$4 WHERE TARGET_COLONY_NAME=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, colonyName, core.RUNNING)
	if err != nil {
		return err
	}

	err = db.DeleteFunctionsByColonyName(colonyName)
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

func (db *PQDatabase) CountExecutorsByColonyName(colonyName string) (int, error) {
	executors, err := db.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return -1, err
	}

	return len(executors), nil
}
