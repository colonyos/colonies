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
			sqlStatement := `UPDATE ` + db.dbPrefix + `EXECUTORS SET EXECUTOR_ID=$1, STATE=$2, COMMISSIONTIME=$3, LASTHEARDFROM=$4, LONG=$5, LAT=$6, LOCDESC=$7, HWMODEL=$8, HWNODES=$9, HWCPU=$10, HWMEM=$11, HWSTORAGE=$12, HWGPUNAME=$13, HWGPUCOUNT=$14, HWGPUNODECOUNT=$15, HWGPUMEM=$16, SWNAME=$17, SWTYPE=$18, SWVERSION=$19, ALLOCATIONS=$20, BLUEPRINT_ID=$21, BLUEPRINT_GEN=$22, HWPLATFORM=$23, HWARCHITECTURE=$24, HWNETWORK=$25 WHERE COLONY_NAME=$26 AND NAME=$27`

			allocationsJSONBytes, err := json.Marshal(executor.Allocations)
			if err != nil {
				return err
			}

			networkJSONBytes, err := json.Marshal(executor.Capabilities.Hardware.Network)
			if err != nil {
				return err
			}

			_, err = db.postgresql.Exec(sqlStatement, executor.ID, core.PENDING, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, executor.Capabilities.Hardware.Model, executor.Capabilities.Hardware.Nodes, executor.Capabilities.Hardware.CPU, executor.Capabilities.Hardware.Memory, executor.Capabilities.Hardware.Storage, executor.Capabilities.Hardware.GPU.Name, executor.Capabilities.Hardware.GPU.Count, executor.Capabilities.Hardware.GPU.NodeCount, executor.Capabilities.Hardware.GPU.Memory, executor.Capabilities.Software.Name, executor.Capabilities.Software.Type, executor.Capabilities.Software.Version, string(allocationsJSONBytes), executor.BlueprintID, executor.BlueprintGen, executor.Capabilities.Hardware.Platform, executor.Capabilities.Hardware.Architecture, string(networkJSONBytes), executor.ColonyName, executor.ColonyName+":"+executor.Name)
			return err
		}
		return errors.New("Executor with name <" + executor.Name + "> already exists in Colony with name <" + executor.ColonyName + ">")
	}

	allocationsJSONBytes, err := json.Marshal(executor.Allocations)
	if err != nil {
		return err
	}

	networkJSONBytes, err := json.Marshal(executor.Capabilities.Hardware.Network)
	if err != nil {
		return err
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `EXECUTORS (NAME, EXECUTOR_TYPE, EXECUTOR_ID, COLONY_NAME, STATE, REQUIRE_FUNC_REG, COMMISSIONTIME, LASTHEARDFROM, LONG, LAT, LOCDESC, HWMODEL, HWNODES, HWCPU, HWMEM, HWSTORAGE, HWGPUNAME, HWGPUCOUNT, HWGPUNODECOUNT, HWGPUMEM, SWNAME, SWTYPE, SWVERSION, ALLOCATIONS, BLUEPRINT_ID, BLUEPRINT_GEN, HWPLATFORM, HWARCHITECTURE, HWNETWORK) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)`
	_, err = db.postgresql.Exec(sqlStatement, executor.ColonyName+":"+executor.Name, executor.Type, executor.ID, executor.ColonyName, 0, executor.RequireFuncReg, time.Now(), executor.LastHeardFromTime, executor.Location.Long, executor.Location.Lat, executor.Location.Description, executor.Capabilities.Hardware.Model, executor.Capabilities.Hardware.Nodes, executor.Capabilities.Hardware.CPU, executor.Capabilities.Hardware.Memory, executor.Capabilities.Hardware.Storage, executor.Capabilities.Hardware.GPU.Name, executor.Capabilities.Hardware.GPU.Count, executor.Capabilities.Hardware.GPU.NodeCount, executor.Capabilities.Hardware.GPU.Memory, executor.Capabilities.Software.Name, executor.Capabilities.Software.Type, executor.Capabilities.Software.Version, string(allocationsJSONBytes), executor.BlueprintID, executor.BlueprintGen, executor.Capabilities.Hardware.Platform, executor.Capabilities.Hardware.Architecture, string(networkJSONBytes))

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
		var allocationsJSONStr string
		var nodeID sql.NullString // kept for backward compatibility with existing database schema (unused)
		var blueprintID sql.NullString
		var blueprintGen sql.NullInt64
		var hwPlatform sql.NullString
		var hwArchitecture sql.NullString
		var hwNetwork sql.NullString

		if err := rows.Scan(&name, &executorType, &id, &colonyName, &state, &requireRunReg, &commissionTime, &lastHeardFromTime, &long, &lat, &desc, &hwModel, &hwNodes, &hwCPU, &hwMem, &hwStorage, &hwGPUName, &hwGPUCount, &hwGPUNodeCount, &hwGPUMem, &swName, &swType, &swVersion, &allocationsJSONStr, &nodeID, &blueprintID, &blueprintGen, &hwPlatform, &hwArchitecture, &hwNetwork); err != nil {
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
		gpu := core.GPU{Name: hwGPUName, Count: hwGPUCount, Memory: hwGPUMem, NodeCount: hwGPUNodeCount}

		// Parse network addresses from JSON
		var networkAddresses []string
		if hwNetwork.Valid && hwNetwork.String != "" {
			json.Unmarshal([]byte(hwNetwork.String), &networkAddresses)
		}

		hw := core.Hardware{
			Model:        hwModel,
			CPU:          hwCPU,
			Memory:       hwMem,
			Storage:      hwStorage,
			GPU:          gpu,
			Nodes:        hwNodes,
			Platform:     hwPlatform.String,
			Architecture: hwArchitecture.String,
			Network:      networkAddresses,
		}
		sw := core.Software{Name: swName, Type: swType, Version: swVersion}
		capabilities := core.Capabilities{Hardware: hw, Software: sw}
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
