package postgresql

import (
	"database/sql"
	"encoding/json"
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

	// If an UNREGISTERED executor exists with the same name, reactivate it instead of creating a new record
	// This supports executors without generation-based naming (e.g., static-named reconcilers)
	if existingExecutor != nil {
		if existingExecutor.State == core.UNREGISTERED {
			// Reactivate the executor by updating its state to PENDING and other fields
			sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET EXECUTOR_ID=$1, STATE=$2, COMMISSIONTIME=$3, LASTHEARDFROM=$4, LONG=$5, LAT=$6, LOCDESC=$7, HARDWARE=$8, SOFTWARE=$9, ALLOCATIONS=$10, BLUEPRINT_ID=$11, BLUEPRINT_GEN=$12 WHERE COLONY_NAME=$13 AND NAME=$14`

			allocationsJSONBytes, err := json.Marshal(executor.Allocations)
			if err != nil {
				return err
			}

			hardwareJSONBytes, err := json.Marshal(executor.Capabilities.Hardware)
			if err != nil {
				return err
			}

			softwareJSONBytes, err := json.Marshal(executor.Capabilities.Software)
			if err != nil {
				return err
			}

			_, err = db.postgresql.Exec(sqlStatement, executor.ID, core.PENDING, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, string(hardwareJSONBytes), string(softwareJSONBytes), string(allocationsJSONBytes), executor.BlueprintID, executor.BlueprintGen, executor.ColonyName, executor.ColonyName+":"+executor.Name)
			return err
		}
		return errors.New("Executor with name <" + executor.Name + "> already exists in Colony with name <" + executor.ColonyName + ">")
	}

	allocationsJSONBytes, err := json.Marshal(executor.Allocations)
	if err != nil {
		return err
	}

	hardwareJSONBytes, err := json.Marshal(executor.Capabilities.Hardware)
	if err != nil {
		return err
	}

	softwareJSONBytes, err := json.Marshal(executor.Capabilities.Software)
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `EXECUTORS (NAME, EXECUTOR_TYPE, EXECUTOR_ID, COLONY_NAME, STATE, REQUIRE_FUNC_REG, COMMISSIONTIME, LASTHEARDFROM, LONG, LAT, LOCDESC, HARDWARE, SOFTWARE, ALLOCATIONS, BLUEPRINT_ID, BLUEPRINT_GEN) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`
	_, err = db.postgresql.Exec(sqlStatement, executor.ColonyName+":"+executor.Name, executor.Type, executor.ID, executor.ColonyName, 0, executor.RequireFuncReg, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, string(hardwareJSONBytes), string(softwareJSONBytes), string(allocationsJSONBytes), executor.BlueprintID, executor.BlueprintGen)

	if err != nil {
		if strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint") {
			return errors.New("Executor not unique, both Name and ExecutorId must be unique within a Colony")
		}
		return err
	}

	return nil
}

func (db *PQDatabase) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	executor, err := db.GetExecutorByName(colonyName, executorName)
	if err != nil {
		return err
	}

	if executor == nil {
		return errors.New("Executor with name <" + executorName + "> does not exists in Colony with name <" + colonyName + ">")
	}

	allocationsJSONBytes, err := json.Marshal(allocations)
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET ALLOCATIONS=$1 WHERE COLONY_NAME=$2 AND NAME=$3`
	_, err = db.postgresql.Exec(sqlStatement, allocationsJSONBytes, colonyName, colonyName+":"+executorName)
	if err != nil {
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
		var hardwareJSONStr sql.NullString
		var softwareJSONStr sql.NullString
		var allocationsJSONStr string
		var nodeID sql.NullString
		var blueprintID sql.NullString
		var blueprintGen sql.NullInt64

		if err := rows.Scan(&name, &executorType, &id, &colonyName, &state, &requireRunReg, &commissionTime, &lastHeardFromTime, &long, &lat, &desc, &hardwareJSONStr, &softwareJSONStr, &allocationsJSONStr, &nodeID, &blueprintID, &blueprintGen); err != nil {
			return nil, err
		}
		_ = nodeID // Intentionally unused - kept for database schema compatibility

		s := strings.Split(name, ":")
		if len(s) != 2 {
			return nil, errors.New("Failed to parse Executor name")
		}
		name = s[1]

		allocations := core.Allocations{}
		err := json.Unmarshal([]byte(allocationsJSONStr), &allocations)
		if err != nil {
			return nil, err
		}

		executor := core.CreateExecutorFromDB(id, executorType, name, colonyName, state, requireRunReg, commissionTime, lastHeardFromTime)
		location := core.Location{Long: long, Lat: lat, Description: desc}
		executor.Location = location

		// Parse hardware array from JSON
		var hardware []core.Hardware
		if hardwareJSONStr.Valid && hardwareJSONStr.String != "" {
			json.Unmarshal([]byte(hardwareJSONStr.String), &hardware)
		}

		// Parse software array from JSON
		var software []core.Software
		if softwareJSONStr.Valid && softwareJSONStr.String != "" {
			json.Unmarshal([]byte(softwareJSONStr.String), &software)
		}

		capabilities := core.Capabilities{Hardware: hardware, Software: software}
		executor.Capabilities = capabilities
		executor.Allocations = allocations
		if blueprintID.Valid {
			executor.BlueprintID = blueprintID.String
		}
		if blueprintGen.Valid {
			executor.BlueprintGen = blueprintGen.Int64
		}

		executors = append(executors, executor)
	}

	return executors, nil
}

func (db *PQDatabase) GetExecutors() ([]*core.Executor, error) {
	// Only return registered executors (exclude unregistered for traceability)
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE STATE!=$1`
	rows, err := db.postgresql.Query(sqlStatement, core.UNREGISTERED)
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
	// Return all executors (filtering happens in CLI layer)
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

func (db *PQDatabase) GetExecutorsByBlueprintID(blueprintID string) ([]*core.Executor, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `EXECUTORS WHERE BLUEPRINT_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, blueprintID)
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

func (db *PQDatabase) ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error {
	sqlStatement := `UPDATE  ` + db.dbPrefix + `EXECUTORS SET EXECUTOR_ID=$1 WHERE COLONY_NAME=$2 AND EXECUTOR_ID=$3`
	_, err := db.postgresql.Query(sqlStatement, newExecutorID, colonyName, oldExecutorID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveExecutorByName(colonyName string, executorName string) error {
	executor, err := db.GetExecutorByName(colonyName, executorName)
	if err != nil {
		return err
	}

	if executor == nil {
		return errors.New("Executor <" + executorName + "> does not exists")
	}

	// Mark executor as unregistered instead of deleting it (for traceability)
	sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET STATE=$1 WHERE COLONY_NAME=$2 AND NAME=$3`
	_, err = db.postgresql.Exec(sqlStatement, core.UNREGISTERED, colonyName, colonyName+":"+executorName)
	if err != nil {
		return err
	}

	// Move back the executor currently running process back to the queue
	sqlStatement = `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_EXECUTOR_ID=$3, STATE=$4 WHERE ASSIGNED_EXECUTOR_ID=$5 AND STATE=$6`
	_, err = db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, executor.ID, core.RUNNING)
	if err != nil {
		return err
	}

	err = db.RemoveFunctionsByExecutorName(executor.ColonyName, executor.Name)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveExecutorsByColonyName(colonyName string) error {
	// When called during colony removal, we permanently delete executors (not just unregister)
	// Move back any currently running processes back to the queue first
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_EXECUTOR_ID=$3, STATE=$4 WHERE TARGET_COLONY_NAME=$5 AND STATE=$6`
	_, err := db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, colonyName, core.RUNNING)
	if err != nil {
		return err
	}

	// Permanently delete all executors in this colony
	sqlStatement = `DELETE FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_NAME=$1`
	_, err = db.postgresql.Exec(sqlStatement, colonyName)
	if err != nil {
		return err
	}

	err = db.RemoveFunctionsByColonyName(colonyName)
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

func (db *PQDatabase) CountExecutorsByColonyNameAndState(colonyName string, state int) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `EXECUTORS WHERE COLONY_NAME=$1 AND STATE=$2`
	var count int
	err := db.postgresql.QueryRow(sqlStatement, colonyName, state).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (db *PQDatabase) UpdateExecutorCapabilities(colonyName string, executorName string, capabilities core.Capabilities) error {
	executor, err := db.GetExecutorByName(colonyName, executorName)
	if err != nil {
		return err
	}

	if executor == nil {
		return errors.New("Executor with name <" + executorName + "> does not exist in Colony with name <" + colonyName + ">")
	}

	hardwareJSONBytes, err := json.Marshal(capabilities.Hardware)
	if err != nil {
		return err
	}

	softwareJSONBytes, err := json.Marshal(capabilities.Software)
	if err != nil {
		return err
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET HARDWARE=$1, SOFTWARE=$2 WHERE COLONY_NAME=$3 AND NAME=$4`
	_, err = db.postgresql.Exec(sqlStatement, string(hardwareJSONBytes), string(softwareJSONBytes), colonyName, colonyName+":"+executorName)
	if err != nil {
		return err
	}

	return nil
}
