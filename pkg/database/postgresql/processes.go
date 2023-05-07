package postgresql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddProcess(process *core.Process) error {
	targetExecutorIDs := process.FunctionSpec.Conditions.ExecutorIDs
	if len(process.FunctionSpec.Conditions.ExecutorIDs) == 0 {
		targetExecutorIDs = []string{"*"}
	}

	submissionTime := time.Now()

	maxWaitTime := process.FunctionSpec.MaxWaitTime
	var deadline time.Time
	if maxWaitTime > 0 {
		deadline = time.Now().Add(time.Duration(maxWaitTime) * time.Second)
	}

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `PROCESSES (PROCESS_ID, TARGET_COLONY_ID, TARGET_EXECUTOR_IDS, ASSIGNED_EXECUTOR_ID, STATE, IS_ASSIGNED, EXECUTOR_TYPE, SUBMISSION_TIME, START_TIME, END_TIME, WAIT_DEADLINE, EXEC_DEADLINE, ERRORS, RETRIES, NODENAME, FUNCNAME, ARGS, MAX_WAIT_TIME, MAX_EXEC_TIME, MAX_RETRIES, DEPENDENCIES, PRIORITY, PRIORITYTIME, WAIT_FOR_PARENTS, PARENTS, CHILDREN, PROCESSGRAPH_ID, INPUT, OUTPUT, LABEL) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)`

	// TODO: Change the database so that argsm input and output are only text
	argsJSON, err := json.Marshal(process.FunctionSpec.Args)
	if err != nil {
		return err
	}
	argsJSONArrStr := []string{string(argsJSON)}

	inJSON, err := json.Marshal(process.Input)
	if err != nil {
		return err
	}
	inJSONArrStr := []string{string(inJSON)}

	outJSON, err := json.Marshal(process.Output)
	if err != nil {
		return err
	}
	outJSONArrStr := []string{string(outJSON)}

	process.SetSubmissionTime(submissionTime)

	_, err = db.postgresql.Exec(sqlStatement, process.ID, process.FunctionSpec.Conditions.ColonyID, pq.Array(targetExecutorIDs), process.AssignedExecutorID, process.State, process.IsAssigned, process.FunctionSpec.Conditions.ExecutorType, submissionTime, time.Time{}, time.Time{}, deadline, process.ExecDeadline, pq.Array(process.Errors), 0, process.FunctionSpec.NodeName, process.FunctionSpec.FuncName, pq.Array(argsJSONArrStr), process.FunctionSpec.MaxWaitTime, process.FunctionSpec.MaxExecTime, process.FunctionSpec.MaxRetries, pq.Array(process.FunctionSpec.Conditions.Dependencies), process.FunctionSpec.Priority, process.PriorityTime, process.WaitForParents, pq.Array(process.Parents), pq.Array(process.Children), process.ProcessGraphID, pq.Array(inJSONArrStr), pq.Array(outJSONArrStr), process.FunctionSpec.Label)
	if err != nil {
		return err
	}

	// Convert Envs to Attributes
	for key, value := range process.FunctionSpec.Env {
		process.Attributes = append(process.Attributes, core.CreateAttribute(process.ID, process.FunctionSpec.Conditions.ColonyID, process.ProcessGraphID, core.ENV, key, value))
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
		var targetColonyID string
		var targetExecutorIDs []string
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
		var argsJSONStrArr []string
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
		var inputJSONStrArr []string
		var outputJSONStrArr []string
		var label string

		if err := rows.Scan(&processID, &targetColonyID, pq.Array(&targetExecutorIDs), &assignedExecutorID, &state, &isAssigned, &executorType, &submissionTime, &startTime, &endTime, &waitDeadline, &execDeadline, pq.Array(&errs), &nodeName, &funcName, pq.Array(&argsJSONStrArr), &maxWaitTime, &maxExecTime, &retries, &maxRetries, pq.Array(&dependencies), &priority, &priorityTime, &waitForParent, pq.Array(&parents), pq.Array(&children), &processGraphID, pq.Array(&inputJSONStrArr), pq.Array(&outputJSONStrArr), &label); err != nil {
			return nil, err
		}

		attributes, err := db.GetAttributes(processID)
		if err != nil {
			return nil, err
		}

		if len(attributes) == 0 {
			attributes = make([]core.Attribute, 0)
		}

		if len(targetExecutorIDs) == 1 && targetExecutorIDs[0] == "*" {
			targetExecutorIDs = []string{}
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
		if len(argsJSONStrArr) == 1 {
			json.Unmarshal([]byte(argsJSONStrArr[0]), &argsif)
		}

		var inputif []interface{}
		if len(inputJSONStrArr) == 1 {
			json.Unmarshal([]byte(inputJSONStrArr[0]), &inputif)
		}

		var outputif []interface{}
		if len(outputJSONStrArr) == 1 {
			json.Unmarshal([]byte(outputJSONStrArr[0]), &outputif)
		}

		functionSpec := core.CreateFunctionSpec(nodeName, funcName, argsif, targetColonyID, targetExecutorIDs, executorType, maxWaitTime, maxExecTime, maxRetries, env, dependencies, priority, label)
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

	if len(processes) > 1 {
		return nil, errors.New("Expected one process, process id should be unique")
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

func (db *PQDatabase) FindProcessesByColonyID(colonyID string, seconds int, state int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 AND SUBMISSION_TIME BETWEEN NOW() - INTERVAL '1 seconds' * $3 AND NOW() ORDER BY SUBMISSION_TIME ASC`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, state, strconv.Itoa(seconds))
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

func (db *PQDatabase) FindProcessesByExecutorID(colonyID string, executorID string, seconds int, state int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND ASSIGNED_EXECUTOR_ID=$2 AND STATE=$3 AND SUBMISSION_TIME BETWEEN NOW() - INTERVAL '1 seconds' * $4 AND NOW() ORDER BY SUBMISSION_TIME ASC`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, executorID, state, strconv.Itoa(seconds))
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

func (db *PQDatabase) FindWaitingProcesses(colonyID string, count int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY PRIORITYTIME LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.WAITING, count)
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

func (db *PQDatabase) FindRunningProcesses(colonyID string, count int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY START_TIME ASC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.RUNNING, count)
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 ORDER BY PRIORITYTIME`
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

func (db *PQDatabase) FindSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY END_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.SUCCESS, count)
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

func (db *PQDatabase) FindFailedProcesses(colonyID string, count int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY END_TIME DESC LIMIT $3`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, core.FAILED, count)
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

func (db *PQDatabase) FindUnassignedProcesses(colonyID string, executorID string, executorType string, count int) ([]*core.Process, error) {
	var sqlStatement string

	sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 AND EXECUTOR_TYPE=$2 AND IS_ASSIGNED=FALSE AND WAIT_FOR_PARENTS=FALSE AND TARGET_COLONY_ID=$3 ORDER BY PRIORITYTIME LIMIT $4`
	rows, err := db.postgresql.Query(sqlStatement, core.WAITING, executorType, colonyID, count)
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

func (db *PQDatabase) DeleteProcessByID(processID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE PROCESS_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processID)
	if err != nil {
		return err
	}

	// TODO test this code
	err = db.DeleteAllAttributesByTargetID(processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllProcesses() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES`
	_, err := db.postgresql.Exec(sqlStatement)
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributes()
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllWaitingProcessesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, "", core.WAITING)
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesByColonyIDWithState(colonyID, core.WAITING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllRunningProcessesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, "", core.RUNNING)
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesByColonyIDWithState(colonyID, core.RUNNING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllSuccessfulProcessesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, "", core.SUCCESS)
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesByColonyIDWithState(colonyID, core.SUCCESS)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllFailedProcessesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID=$2 AND STATE=$3`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, "", core.FAILED)
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesByColonyIDWithState(colonyID, core.FAILED)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllProcessesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, "")
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesByColonyID(colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllProcessesByProcessGraphID(processGraphID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE PROCESSGRAPH_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, processGraphID)
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesByProcessGraphID(processGraphID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllProcessesInProcessGraphsByColonyID(colonyID string) error {
	err := db.DeleteAllAttributesInProcessGraphsByColonyID(colonyID)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID!=$2`
	_, err = db.postgresql.Exec(sqlStatement, colonyID, "")
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID string, state int) error {
	err := db.DeleteAllAttributesInProcessGraphsByColonyIDWithState(colonyID, state)
	if err != nil {
		return err
	}

	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID!=$2 AND STATE=$3`
	_, err = db.postgresql.Exec(sqlStatement, colonyID, "", state)
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
	inJSONArrStr := []string{string(inJSON)}

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET INPUT=$1 WHERE PROCESS_ID=$2`
	_, err = db.postgresql.Exec(sqlStatement, pq.Array(inJSONArrStr), processID)
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
	outJSONArrStr := []string{string(outJSON)}

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET OUTPUT=$1 WHERE PROCESS_ID=$2`
	_, err = db.postgresql.Exec(sqlStatement, pq.Array(outJSONArrStr), processID)
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
		return 0.0, 0.0, errors.New("Tried to set failed process as completed")
	}

	if process.State == core.WAITING {
		return 0.0, 0.0, errors.New("Tried to set waiting process as completed without being running")
	}

	if process.State == core.FAILED {
		return 0.0, 0.0, errors.New("Tried to set failed process (from db) as successful")
	}

	if process.State == core.WAITING {
		return 0.0, 0.0, errors.New("Tried to set waiting process (from db) as successful without being running")
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

	if process.State == core.SUCCESS {
		return errors.New("Tried to set successful process as failed")
	}

	if process.State == core.FAILED {
		return errors.New("Tried to set failed process as failed")
	}

	if process.State == core.SUCCESS {
		return errors.New("Tried to set successful (from db) as failed")
	}

	if process.State == core.FAILED {
		return errors.New("Tried to set failed (from db) as failed")
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

func (db *PQDatabase) countProcessesByColonyID(state int, colonyID string) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 AND TARGET_COLONY_ID=$2`
	rows, err := db.postgresql.Query(sqlStatement, state, colonyID)
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

func (db *PQDatabase) CountWaitingProcessesByColonyID(colonyID string) (int, error) {
	return db.countProcessesByColonyID(core.WAITING, colonyID)
}

func (db *PQDatabase) CountRunningProcessesByColonyID(colonyID string) (int, error) {
	return db.countProcessesByColonyID(core.RUNNING, colonyID)
}

func (db *PQDatabase) CountSuccessfulProcessesByColonyID(colonyID string) (int, error) {
	return db.countProcessesByColonyID(core.SUCCESS, colonyID)
}

func (db *PQDatabase) CountFailedProcessesByColonyID(colonyID string) (int, error) {
	return db.countProcessesByColonyID(core.FAILED, colonyID)
}
