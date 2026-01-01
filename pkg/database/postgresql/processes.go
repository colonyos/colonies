package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/parsers"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddProcess(process *core.Process) error {
	targetExecutorNames := process.FunctionSpec.Conditions.ExecutorNames

	submissionTime := time.Now()

	maxWaitTime := process.FunctionSpec.MaxWaitTime
	var deadline time.Time
	if maxWaitTime > 0 {
		deadline = time.Now().Add(time.Duration(maxWaitTime) * time.Second)
	}

	fsJSONStr, err := json.Marshal(process.FunctionSpec.Filesystem)
	if err != nil {
		return nil
	}

	// Blueprint field removed from FunctionSpec - always write empty string for column
	blueprintJSONStr := ""

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `PROCESSES (PROCESS_ID, TARGET_COLONY_NAME, TARGET_EXECUTOR_NAMES, ASSIGNED_EXECUTOR_ID, STATE, IS_ASSIGNED, EXECUTOR_TYPE, SUBMISSION_TIME, START_TIME, END_TIME, WAIT_DEADLINE, EXEC_DEADLINE, ERRORS, RETRIES, NODENAME, FUNCNAME, ARGS, KWARGS, MAX_WAIT_TIME, MAX_EXEC_TIME, MAX_RETRIES, DEPENDENCIES, PRIORITY, PRIORITYTIME, WAIT_FOR_PARENTS, PARENTS, CHILDREN, PROCESSGRAPH_ID, INPUT, OUTPUT, LABEL, FS, NODES, CPU, PROCESSES, PROCESSES_PER_NODE, MEMORY, STORAGE, GPUNAME, GPUCOUNT, GPUMEM, WALLTIME, INITIATOR_ID, INITIATOR_NAME, BLUEPRINT, CHANNELS, LOCATION_NAME) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47)`

	argsJSON, err := json.Marshal(process.FunctionSpec.Args)
	if err != nil {
		return err
	}
	argsJSONStr := string(argsJSON)

	kwargsJSON, err := json.Marshal(process.FunctionSpec.KwArgs)
	if err != nil {
		return err
	}
	kwargsJSONStr := string(kwargsJSON)

	inJSON, err := json.Marshal(process.Input)
	if err != nil {
		return err
	}
	inJSONStr := string(inJSON)

	outJSON, err := json.Marshal(process.Output)
	if err != nil {
		return err
	}
	outJSONStr := string(outJSON)

	process.SetSubmissionTime(submissionTime)
	process.WaitDeadline = deadline

	cpu, err := parsers.ConvertCPUToInt(process.FunctionSpec.Conditions.CPU)
	if err != nil {
		return err
	}

	memory, err := parsers.ConvertMemoryToBytes(process.FunctionSpec.Conditions.Memory)
	if err != nil {
		return err
	}

	storage, err := parsers.ConvertMemoryToBytes(process.FunctionSpec.Conditions.Storage)
	if err != nil {
		return err
	}

	gpuMem, err := parsers.ConvertMemoryToBytes(process.FunctionSpec.Conditions.GPU.Memory)
	if err != nil {
		return err
	}

	_, err = db.postgresql.Exec(sqlStatement, process.ID, process.FunctionSpec.Conditions.ColonyName, pq.Array(targetExecutorNames), process.AssignedExecutorID, process.State, process.IsAssigned, process.FunctionSpec.Conditions.ExecutorType, submissionTime, time.Time{}, time.Time{}, deadline, process.ExecDeadline, pq.Array(process.Errors), 0, process.FunctionSpec.NodeName, process.FunctionSpec.FuncName, argsJSONStr, kwargsJSONStr, process.FunctionSpec.MaxWaitTime, process.FunctionSpec.MaxExecTime, process.FunctionSpec.MaxRetries, pq.Array(process.FunctionSpec.Conditions.Dependencies), process.FunctionSpec.Priority, process.PriorityTime, process.WaitForParents, pq.Array(process.Parents), pq.Array(process.Children), process.ProcessGraphID, inJSONStr, outJSONStr, process.FunctionSpec.Label, fsJSONStr, process.FunctionSpec.Conditions.Nodes, cpu, process.FunctionSpec.Conditions.Processes, process.FunctionSpec.Conditions.ProcessesPerNode, memory, storage, process.FunctionSpec.Conditions.GPU.Name, process.FunctionSpec.Conditions.GPU.Count, gpuMem, process.FunctionSpec.Conditions.WallTime, process.InitiatorID, process.InitiatorName, blueprintJSONStr, pq.Array(process.FunctionSpec.Channels), process.FunctionSpec.Conditions.LocationName)
	if err != nil {
		return err
	}

	// Convert Envs to Attributes
	for key, value := range process.FunctionSpec.Env {
		process.Attributes = append(process.Attributes, core.CreateAttribute(process.ID, process.FunctionSpec.Conditions.ColonyName, process.ProcessGraphID, core.ENV, key, value))
	}

	err = db.AddAttributes(process.Attributes)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) parseProcesses(rows *sql.Rows) ([]*core.Process, error) {
	var processes []*core.Process

	for rows.Next() {
		var processID string
		var targetColonyName string
		var targetExecutorNames []string
		var assignedExecutorID string
		var state int
		var isAssigned bool
		var executorType string
		var submissionTime time.Time
		var startTime time.Time
		var endTime time.Time
		var waitDeadline time.Time
		var execDeadline time.Time
		var errs []string
		var nodeName string
		var funcName string
		var argsJSONStr string
		var kwargsJSONStr string
		var maxWaitTime int
		var maxExecTime int
		var retries int
		var maxRetries int
		var dependencies []string
		var priority int
		var priorityTime int64
		var waitForParent bool
		var parents []string
		var children []string
		var processGraphID string
		var inputJSONStr string
		var outputJSONStr string
		var label string
		var fsJSONStr string
		var nodes int
		var cpu int64
		var processesCount int
		var processesPerNode int
		var memory int64
		var storage int64
		var gpuName string
		var gpuCount int
		var gpuMemory int64
		var walltime int64
		var initiatorID string
		var initiatorName string
		var blueprintJSONStr sql.NullString
		var channels []string
		var locationName sql.NullString

		if err := rows.Scan(&processID, &targetColonyName, pq.Array(&targetExecutorNames), &assignedExecutorID, &state, &isAssigned, &executorType, &submissionTime, &startTime, &endTime, &waitDeadline, &execDeadline, pq.Array(&errs), &nodeName, &funcName, &argsJSONStr, &kwargsJSONStr, &maxWaitTime, &maxExecTime, &retries, &maxRetries, pq.Array(&dependencies), &priority, &priorityTime, &waitForParent, pq.Array(&parents), pq.Array(&children), &processGraphID, &inputJSONStr, &outputJSONStr, &label, &fsJSONStr, &nodes, &cpu, &processesCount, &processesPerNode, &memory, &storage, &gpuName, &gpuCount, &gpuMemory, &walltime, &initiatorID, &initiatorName, &blueprintJSONStr, pq.Array(&channels), &locationName); err != nil {
			return nil, err
		}

		attributes, err := db.GetAttributes(processID)
		if err != nil {
			return nil, err
		}

		if len(attributes) == 0 {
			attributes = make([]core.Attribute, 0)
		}

		if len(targetExecutorNames) == 1 && targetExecutorNames[0] == "*" {
			targetExecutorNames = []string{}
		}

		// Restore env map
		env := make(map[string]string)
		inAttributes, err := db.GetAttributesByType(processID, core.ENV)
		if err != nil {
			return nil, err
		}

		for _, attribute := range inAttributes {
			env[attribute.Key] = attribute.Value
		}

		if len(dependencies) == 0 {
			dependencies = make([]string, 0)
		}

		var argsif []interface{}
		err = json.Unmarshal([]byte(argsJSONStr), &argsif)
		if err != nil {
			return nil, err
		}

		var kwargsif map[string]interface{}
		err = json.Unmarshal([]byte(kwargsJSONStr), &kwargsif)
		if err != nil {
			return nil, err
		}

		var inputif []interface{}
		err = json.Unmarshal([]byte(inputJSONStr), &inputif)
		if err != nil {
			return nil, err
		}

		var outputif []interface{}
		err = json.Unmarshal([]byte(outputJSONStr), &outputif)
		if err != nil {
			return nil, err
		}

		functionSpec := core.CreateFunctionSpec(nodeName, funcName, argsif, kwargsif, targetColonyName, targetExecutorNames, executorType, maxWaitTime, maxExecTime, maxRetries, env, dependencies, priority, label)

		// Set channels
		if len(channels) == 0 {
			functionSpec.Channels = make([]string, 0)
		} else {
			functionSpec.Channels = channels
		}

		functionSpec.Conditions.Nodes = nodes
		functionSpec.Conditions.CPU = parsers.ConvertCPUToString(cpu)
		functionSpec.Conditions.Processes = processesCount
		functionSpec.Conditions.ProcessesPerNode = processesPerNode
		functionSpec.Conditions.Memory = parsers.ConvertMemoryToString(memory)
		functionSpec.Conditions.Storage = parsers.ConvertMemoryToString(storage)
		functionSpec.Conditions.GPU.Name = gpuName
		functionSpec.Conditions.GPU.Count = gpuCount
		functionSpec.Conditions.GPU.Memory = parsers.ConvertMemoryToString(gpuMemory)
		functionSpec.Conditions.WallTime = walltime
		if locationName.Valid {
			functionSpec.Conditions.LocationName = locationName.String
		}

		fs := core.Filesystem{}
		err = json.Unmarshal([]byte(fsJSONStr), &fs)
		if err != nil {
			return nil, err
		}
		functionSpec.Filesystem = fs

		// Blueprint field removed from FunctionSpec - skip deserialization
		// Column still exists in DB for backwards compatibility but is no longer used

		process := core.CreateProcessFromDB(functionSpec, processID, assignedExecutorID, isAssigned, state, priorityTime, submissionTime, startTime, endTime, waitDeadline, execDeadline, errs, retries, attributes)

		process.Input = inputif
		process.Output = outputif
		processes = append(processes, process)

		process.WaitForParents = waitForParent
		if len(parents) == 0 {
			process.Parents = make([]string, 0)
		} else {
			process.Parents = parents
		}
		if len(children) == 0 {
			process.Children = make([]string, 0)
		} else {
			process.Children = children
		}
		process.ProcessGraphID = processGraphID
		process.InitiatorID = initiatorID
		process.InitiatorName = initiatorName
	}

	return processes, nil
}

func (db *PQDatabase) GetProcesses() ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES ORDER BY SUBMISSION_TIME DESC`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseProcesses(rows)
}

func (db *PQDatabase) GetProcessByID(processID string) (*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE PROCESS_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, processID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	processes, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	if len(processes) == 0 {
		return nil, nil
	}

	return processes[0], nil
}

func (db *PQDatabase) selectCandidate(candidates []*core.Process) *core.Process {
	if len(candidates) > 0 {
		return candidates[0]
	} else {
		return nil
	}
}

func (db *PQDatabase) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 AND SUBMISSION_TIME BETWEEN NOW() - INTERVAL '1 seconds' * $3 AND NOW() ORDER BY SUBMISSION_TIME ASC`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, state, strconv.Itoa(seconds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND ASSIGNED_EXECUTOR_ID=$2 AND STATE=$3 AND SUBMISSION_TIME BETWEEN NOW() - INTERVAL '1 seconds' * $4 AND NOW() ORDER BY SUBMISSION_TIME ASC`
	rows, err := db.postgresql.Query(sqlStatement, colonyName, executorID, state, strconv.Itoa(seconds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	var sqlStatement string
	var rows *sql.Rows
	var err error

	if executorType != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND EXECUTOR_TYPE=$2 AND STATE=$3 ORDER BY PRIORITYTIME LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, executorType, core.WAITING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if initiator != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND INITIATOR_NAME=$2 AND STATE=$3 ORDER BY PRIORITYTIME LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, initiator, core.WAITING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if label != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND LABEL=$2 AND STATE=$3 ORDER BY PRIORITYTIME LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, label, core.WAITING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 ORDER BY PRIORITYTIME LIMIT $3`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, core.WAITING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	}

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	var sqlStatement string
	var rows *sql.Rows
	var err error

	if executorType != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND EXECUTOR_TYPE=$2 AND STATE=$3 ORDER BY START_TIME ASC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, executorType, core.RUNNING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if initiator != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND INITIATOR_NAME=$2 AND STATE=$3 ORDER BY START_TIME ASC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, initiator, core.RUNNING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if label != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND LABEL=$2 AND STATE=$3 ORDER BY START_TIME ASC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, label, core.RUNNING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 ORDER BY START_TIME ASC LIMIT $3`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, core.RUNNING, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	}

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	var sqlStatement string
	var rows *sql.Rows
	var err error

	if executorType != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND EXECUTOR_TYPE=$2 AND STATE=$3 ORDER BY END_TIME DESC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, executorType, core.SUCCESS, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if initiator != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND INITIATOR_NAME=$2 AND STATE=$3 ORDER BY END_TIME DESC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, initiator, core.SUCCESS, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if label != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND LABEL=$2 AND STATE=$3 ORDER BY END_TIME DESC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, label, core.SUCCESS, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 ORDER BY END_TIME DESC LIMIT $3`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, core.SUCCESS, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	}

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	var sqlStatement string
	var rows *sql.Rows
	var err error

	if executorType != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND EXECUTOR_TYPE=$2 AND STATE=$3 ORDER BY END_TIME DESC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, executorType, core.FAILED, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if initiator != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND INITIATOR_NAME=$2 AND STATE=$3 ORDER BY END_TIME DESC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, initiator, core.FAILED, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else if label != "" {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND LABEL=$2 AND STATE=$3 ORDER BY END_TIME DESC LIMIT $4`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, label, core.FAILED, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND STATE=$2 ORDER BY END_TIME DESC LIMIT $3`
		rows, err = db.postgresql.Query(sqlStatement, colonyName, core.FAILED, count)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	}

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindAllRunningProcesses() ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 ORDER BY START_TIME ASC`
	rows, err := db.postgresql.Query(sqlStatement, core.RUNNING)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindAllWaitingProcesses() ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 ORDER BY PRIORITYTIME LIMIT 1000`
	rows, err := db.postgresql.Query(sqlStatement, core.WAITING)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	var sqlStatement string

	sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 AND EXECUTOR_TYPE=$2 AND IS_ASSIGNED=FALSE AND WAIT_FOR_PARENTS=FALSE AND TARGET_COLONY_NAME=$3 AND array_length(TARGET_EXECUTOR_NAMES, 1) IS NULL AND CPU<=$4 AND MEMORY<=$5 AND STORAGE<=$6 AND NODES<=$7 AND PROCESSES<=$8 AND PROCESSES_PER_NODE<=$9 AND (LOCATION_NAME IS NULL OR LOCATION_NAME = '' OR LOWER(LOCATION_NAME) = LOWER($10)) ORDER BY PRIORITYTIME LIMIT $11`
	rows, err := db.postgresql.Query(sqlStatement, core.WAITING, executorType, colonyName, cpu, memory, storage, nodes, processes, processesPerNode, executorLocationName, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	var sqlStatement string

	sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 AND $2=ANY(TARGET_EXECUTOR_NAMES) AND EXECUTOR_TYPE=$3 AND IS_ASSIGNED=FALSE AND WAIT_FOR_PARENTS=FALSE AND TARGET_COLONY_NAME=$4 AND CPU<=$5 AND MEMORY<=$6 AND STORAGE<=$7 AND NODES<=$8 AND PROCESSES<=$9 AND PROCESSES_PER_NODE<=$10 AND (LOCATION_NAME IS NULL OR LOCATION_NAME = '' OR LOWER(LOCATION_NAME) = LOWER($11)) ORDER BY PRIORITYTIME LIMIT $12`
	rows, err := db.postgresql.Query(sqlStatement, core.WAITING, executorName, executorType, colonyName, cpu, memory, storage, nodes, processes, processesPerNode, executorLocationName, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func (db *PQDatabase) RemoveProcessByID(processID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE PROCESS_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processID)
	if err != nil {
		return err
	}

	// TODO test this code
	err = db.RemoveAllAttributesByTargetID(processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllProcesses() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributes()
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllWaitingProcessesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, "", core.WAITING)
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.WAITING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllRunningProcessesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, "", core.RUNNING)
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, "", core.SUCCESS)
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.SUCCESS)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllFailedProcessesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, "", core.FAILED)
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributesByColonyNameWithState(colonyName, core.FAILED)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllProcessesByColonyName(colonyName string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyName, "")
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributesByColonyName(colonyName)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllProcessesByProcessGraphID(processGraphID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE PROCESSGRAPH_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processGraphID)
	if err != nil {
		return err
	}

	err = db.RemoveAllAttributesByProcessGraphID(processGraphID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error {
	err := db.RemoveAllAttributesInProcessGraphsByColonyName(colonyName)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID!=$2`
	_, err = db.postgresql.Exec(sqlStatement, colonyName, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) RemoveAllProcessesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	err := db.RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName, state)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_NAME=$1 AND PROCESSGRAPH_ID!=$2 AND STATE=$3`
	_, err = db.postgresql.Exec(sqlStatement, colonyName, "", state)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) ResetProcess(process *core.Process) error {
	submissionTime := time.Now()

	maxWaitTime := process.FunctionSpec.MaxWaitTime
	if maxWaitTime > 0 {
		deadline := time.Now().Add(time.Duration(maxWaitTime) * time.Second)
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, SUBMISSION_TIME=$1, START_TIME=$2, END_TIME=$3, ASSIGNED_EXECUTOR_ID=$4, STATE=$5, WAIT_DEADLINE=$6 WHERE PROCESS_ID=$7`
		_, err := db.postgresql.Exec(sqlStatement, submissionTime, time.Time{}, time.Time{}, "", core.WAITING, deadline, process.ID)
		if err != nil {
			return err
		}
	} else {
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, SUBMISSION_TIME=$1, START_TIME=$2, END_TIME=$3, ASSIGNED_EXECUTOR_ID=$4, STATE=$5 WHERE PROCESS_ID=$6`
		_, err := db.postgresql.Exec(sqlStatement, submissionTime, time.Time{}, time.Time{}, "", core.WAITING, process.ID)
		if err != nil {
			return err
		}
	}

	process.SetSubmissionTime(submissionTime)
	process.SetStartTime(time.Time{})
	process.SetEndTime(time.Time{})
	process.SetAssignedExecutorID("")
	process.SetState(core.WAITING)

	return nil
}

func (db *PQDatabase) SetWaitForParents(processID string, waitForParent bool) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET WAIT_FOR_PARENTS=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, waitForParent, processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetParents(processID string, parents []string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET PARENTS=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, pq.Array(parents), processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetChildren(processID string, children []string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET CHILDREN=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, pq.Array(children), processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetProcessState(processID string, state int) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET STATE=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, state, processID)
	if err != nil {
		return err
	}

	err = db.SetAttributeState(processID, state)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetInput(processID string, input []interface{}) error {
	inJSON, err := json.Marshal(input)
	if err != nil {
		return err
	}
	inJSONStr := string(inJSON)

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET INPUT=$1 WHERE PROCESS_ID=$2`
	_, err = db.postgresql.Exec(sqlStatement, inJSONStr, processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetOutput(processID string, output []interface{}) error {
	outJSON, err := json.Marshal(output)
	if err != nil {
		return err
	}
	outJSONStr := string(outJSON)

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET OUTPUT=$1 WHERE PROCESS_ID=$2`
	_, err = db.postgresql.Exec(sqlStatement, outJSONStr, processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetErrors(processID string, errs []string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET ERRORS=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, pq.Array(errs), processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetExecDeadline(process *core.Process, execDeadline time.Time) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET EXEC_DEADLINE=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, execDeadline, process.ID)
	if err != nil {
		return err
	}

	process.ExecDeadline = execDeadline

	return nil
}

func (db *PQDatabase) SetWaitDeadline(process *core.Process, waitDeadline time.Time) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET WAIT_DEADLINE=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, waitDeadline, process.ID)
	if err != nil {
		return err
	}

	process.ExecDeadline = waitDeadline

	return nil
}

func (db *PQDatabase) Assign(executorID string, process *core.Process) error {
	processFromDB, err := db.GetProcessByID(process.ID)
	if err != nil {
		return err
	}

	if processFromDB.IsAssigned {
		return errors.New("Process already assigned")
	}

	startTime := time.Now()
	if process.FunctionSpec.MaxExecTime > 0 {
		deadline := time.Now().Add(time.Duration(process.FunctionSpec.MaxExecTime) * time.Second)
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=TRUE, START_TIME=$1, ASSIGNED_EXECUTOR_ID=$2, STATE=$3, EXEC_DEADLINE=$4 WHERE PROCESS_ID=$5`
		_, err = db.postgresql.Exec(sqlStatement, startTime, executorID, core.RUNNING, deadline, process.ID)
		if err != nil {
			return err
		}
	} else {
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=TRUE, START_TIME=$1, ASSIGNED_EXECUTOR_ID=$2, STATE=$3 WHERE PROCESS_ID=$4`
		_, err = db.postgresql.Exec(sqlStatement, startTime, executorID, core.RUNNING, process.ID)
		if err != nil {
			return err
		}
	}

	err = db.SetAttributeState(process.ID, core.RUNNING)
	if err != nil {
		return err
	}

	process.SetStartTime(startTime)
	process.Assign()
	process.SetAssignedExecutorID(executorID)
	process.SetState(core.RUNNING)

	return nil
}

// SelectAndAssign atomically selects a candidate process and assigns it to the executor.
// Uses FOR UPDATE SKIP LOCKED to handle concurrent access without race conditions.
// This enables distributed assignment across multiple server replicas.
func (db *PQDatabase) SelectAndAssign(colonyName string, executorID string, executorName string, executorType string, executorLocation string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) (*core.Process, error) {
	// Atomic SELECT FOR UPDATE SKIP LOCKED + UPDATE in a single statement
	// The subquery locks the row, preventing other transactions from selecting it.
	// Uses OR to combine both FindCandidatesByName and FindCandidates logic:
	// - Matches processes targeting this specific executor by name
	// - OR matches general pool processes (no specific executor names)
	sqlStatement := `
		UPDATE ` + db.dbPrefix + `PROCESSES
		SET IS_ASSIGNED = TRUE,
		    START_TIME = NOW(),
		    ASSIGNED_EXECUTOR_ID = $12,
		    STATE = $13,
		    EXEC_DEADLINE = CASE
		        WHEN MAX_EXEC_TIME > 0 THEN NOW() + (MAX_EXEC_TIME * INTERVAL '1 second')
		        ELSE EXEC_DEADLINE
		    END
		WHERE PROCESS_ID = (
			SELECT PROCESS_ID FROM ` + db.dbPrefix + `PROCESSES
			WHERE (
				-- By executor name (processes targeting this specific executor)
				$1 = ANY(TARGET_EXECUTOR_NAMES)
				OR
				-- By executor type (general pool - processes not targeting specific executors)
				(array_length(TARGET_EXECUTOR_NAMES, 1) IS NULL OR TARGET_EXECUTOR_NAMES = ARRAY['*']::text[])
			)
			  AND EXECUTOR_TYPE = $2
			  AND STATE = $3
			  AND IS_ASSIGNED = FALSE
			  AND WAIT_FOR_PARENTS = FALSE
			  AND TARGET_COLONY_NAME = $4
			  AND CPU <= $5 AND MEMORY <= $6 AND STORAGE <= $7
			  AND NODES <= $8 AND PROCESSES <= $9 AND PROCESSES_PER_NODE <= $10
			  AND (LOCATION_NAME IS NULL OR LOCATION_NAME = '' OR LOWER(LOCATION_NAME) = LOWER($11))
			ORDER BY PRIORITYTIME ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING *
	`

	rows, err := db.postgresql.Query(sqlStatement,
		executorName,     // $1
		executorType,     // $2
		core.WAITING,     // $3
		colonyName,       // $4
		cpu,              // $5
		memory,           // $6
		storage,          // $7
		nodes,            // $8
		processes,        // $9
		processesPerNode, // $10
		executorLocation, // $11
		executorID,       // $12
		core.RUNNING,     // $13
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	selectedProcesses, err := db.parseProcesses(rows)
	if err != nil {
		return nil, err
	}

	if len(selectedProcesses) == 0 {
		// No process could be selected (either none available or all locked by other transactions)
		return nil, nil
	}

	process := selectedProcesses[0]

	// Update attribute state
	err = db.SetAttributeState(process.ID, core.RUNNING)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func (db *PQDatabase) Unassign(process *core.Process) error {
	endTime := time.Now()

	maxWaitTime := process.FunctionSpec.MaxWaitTime
	if maxWaitTime > 0 {
		deadline := time.Now().Add(time.Duration(maxWaitTime) * time.Second)

		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, END_TIME=$1, STATE=$2, RETRIES=$3, ASSIGNED_EXECUTOR_ID=$4, WAIT_DEADLINE=$5 WHERE PROCESS_ID=$6`
		_, err := db.postgresql.Exec(sqlStatement, endTime, core.WAITING, process.Retries+1, "", deadline, process.ID)
		if err != nil {
			return err
		}
	} else {
		sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, END_TIME=$1, STATE=$2, RETRIES=$3, ASSIGNED_EXECUTOR_ID=$4 WHERE PROCESS_ID=$5`
		_, err := db.postgresql.Exec(sqlStatement, endTime, core.WAITING, process.Retries+1, "", process.ID)
		if err != nil {
			return err
		}
	}

	err := db.SetAttributeState(process.ID, core.PENDING)
	if err != nil {
		return err
	}

	process.SetEndTime(endTime)
	process.Unassign()
	process.SetState(core.WAITING)

	return nil
}

func (db *PQDatabase) MarkSuccessful(processID string) (float64, float64, error) {
	process, err := db.GetProcessByID(processID)
	if err != nil {
		return 0.0, 0.0, err
	}

	if process.State == core.FAILED {
		return 0.0, 0.0, errors.New("Tried to set failed process as successful")
	}

	if process.State == core.WAITING {
		return 0.0, 0.0, errors.New("Tried to set waiting process as successful without being running")
	}

	endTime := time.Now()

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET END_TIME=$1, STATE=$2 WHERE PROCESS_ID=$3`
	_, err = db.postgresql.Exec(sqlStatement, endTime, core.SUCCESS, process.ID)
	if err != nil {
		return 0.0, 0.0, err
	}

	err = db.SetAttributeState(process.ID, core.SUCCESS)
	if err != nil {
		return 0.0, 0.0, err
	}

	process.SetEndTime(endTime)
	process.SetState(core.SUCCESS)

	return process.WaitingTime().Seconds(), process.ProcessingTime().Seconds(), nil
}

func (db *PQDatabase) MarkFailed(processID string, errs []string) error {
	endTime := time.Now()
	process, err := db.GetProcessByID(processID)
	if err != nil {
		return err
	}

	if process == nil {
		return errors.New("Process with Id <" + processID + "> not found")
	}

	if process.State == core.SUCCESS {
		return errors.New("Tried to set successful process as failed")
	}

	if process.State == core.FAILED {
		return errors.New("Tried to set failed process as failed")
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET END_TIME=$1, STATE=$2 WHERE PROCESS_ID=$3`
	_, err = db.postgresql.Exec(sqlStatement, endTime, core.FAILED, process.ID)
	if err != nil {
		return err
	}

	err = db.SetAttributeState(process.ID, core.FAILED)
	if err != nil {
		return err
	}

	process.SetEndTime(endTime)
	process.SetState(core.FAILED)

	return db.SetErrors(process.ID, errs)
}

func (db *PQDatabase) CountProcesses() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `PROCESSES`
	rows, err := db.postgresql.Query(sqlStatement)
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

func (db *PQDatabase) countProcesses(state int) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1`
	rows, err := db.postgresql.Query(sqlStatement, state)
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

// TODO: may be switch to pg_class to improve count performance?
//
// The basic SQL standard query to count the rows in a table is:
// SELECT count(*) FROM table_name;
// This can be rather slow because PostgreSQL has to check visibility for all rows, due to the MVCC model.
// If you don't need an exact count, the current statistic from the catalog table pg_class might be good enough and is much faster to   retrieve for big tables.
// SELECT reltuples AS estimate FROM pg_class WHERE relname = 'table_name';
//
// https://wiki.postgresql.org/wiki/Count_estimate

func (db *PQDatabase) countProcessesByColonyName(state int, colonyName string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 AND TARGET_COLONY_NAME=$2`
	rows, err := db.postgresql.Query(sqlStatement, state, colonyName)
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

func (db *PQDatabase) CountWaitingProcesses() (int, error) {
	return db.countProcesses(core.WAITING)
}

func (db *PQDatabase) CountRunningProcesses() (int, error) {
	return db.countProcesses(core.RUNNING)
}

func (db *PQDatabase) CountSuccessfulProcesses() (int, error) {
	return db.countProcesses(core.SUCCESS)
}

func (db *PQDatabase) CountFailedProcesses() (int, error) {
	return db.countProcesses(core.FAILED)
}

func (db *PQDatabase) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	return db.countProcessesByColonyName(core.WAITING, colonyName)
}

func (db *PQDatabase) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	return db.countProcessesByColonyName(core.RUNNING, colonyName)
}

func (db *PQDatabase) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	return db.countProcessesByColonyName(core.SUCCESS, colonyName)
}

func (db *PQDatabase) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	return db.countProcessesByColonyName(core.FAILED, colonyName)
}
