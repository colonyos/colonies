package postgresql

import (
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddProcess(process *core.Process) error {
	targetRuntimeIDs := process.ProcessSpec.Conditions.RuntimeIDs
	if len(process.ProcessSpec.Conditions.RuntimeIDs) == 0 {
		targetRuntimeIDs = []string{"*"}
	}

	submissionTime := time.Now()

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `PROCESSES (PROCESS_ID, TARGET_COLONY_ID, TARGET_RUNTIME_IDS, ASSIGNED_RUNTIME_ID, STATE, IS_ASSIGNED, RUNTIME_TYPE, SUBMISSION_TIME, START_TIME, END_TIME, WAIT_DEADLINE, EXEC_DEADLINE, ERRORS, RETRIES, NAME, FUNC, ARGS, MAX_WAIT_TIME, MAX_EXEC_TIME, MAX_RETRIES, DEPENDENCIES, PRIORITY, WAIT_FOR_PARENTS, PARENTS, CHILDREN, PROCESSGRAPH_ID, RESULTS) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)`
	_, err := db.postgresql.Exec(sqlStatement, process.ID, process.ProcessSpec.Conditions.ColonyID, pq.Array(targetRuntimeIDs), process.AssignedRuntimeID, process.State, process.IsAssigned, process.ProcessSpec.Conditions.RuntimeType, submissionTime, time.Time{}, time.Time{}, process.WaitDeadline, process.ExecDeadline, pq.Array(process.Errors), 0, process.ProcessSpec.Name, process.ProcessSpec.Func, pq.Array(process.ProcessSpec.Args), process.ProcessSpec.MaxWaitTime, process.ProcessSpec.MaxExecTime, process.ProcessSpec.MaxRetries, pq.Array(process.ProcessSpec.Conditions.Dependencies), process.ProcessSpec.Priority, process.WaitForParents, pq.Array(process.Parents), pq.Array(process.Children), process.ProcessGraphID, pq.Array(process.Results))
	if err != nil {
		return err
	}

	// Convert Envs to Attributes
	for key, value := range process.ProcessSpec.Env {
		process.Attributes = append(process.Attributes, core.CreateAttribute(process.ID, process.ProcessSpec.Conditions.ColonyID, process.ProcessGraphID, core.ENV, key, value))
	}

	err = db.AddAttributes(process.Attributes)
	if err != nil {
		return err
	}

	process.SetSubmissionTime(submissionTime)

	return nil
}

func (db *PQDatabase) parseProcesses(rows *sql.Rows) ([]*core.Process, error) {
	var processes []*core.Process

	for rows.Next() {
		var processID string
		var targetColonyID string
		var targetRuntimeIDs []string
		var assignedRuntimeID string
		var state int
		var isAssigned bool
		var runtimeType string
		var submissionTime time.Time
		var startTime time.Time
		var endTime time.Time
		var waitDeadline time.Time
		var execDeadline time.Time
		var errs []string
		var name string
		var fn string
		var args []string
		var maxWaitTime int
		var maxExecTime int
		var retries int
		var maxRetries int
		var dependencies []string
		var priority int
		var waitForParent bool
		var parents []string
		var children []string
		var processGraphID string
		var results []string

		if err := rows.Scan(&processID, &targetColonyID, pq.Array(&targetRuntimeIDs), &assignedRuntimeID, &state, &isAssigned, &runtimeType, &submissionTime, &startTime, &endTime, &waitDeadline, &execDeadline, pq.Array(&errs), &name, &fn, pq.Array(&args), &maxWaitTime, &maxExecTime, &retries, &maxRetries, pq.Array(&dependencies), &priority, &waitForParent, pq.Array(&parents), pq.Array(&children), &processGraphID, pq.Array(&results)); err != nil {
			return nil, err
		}

		attributes, err := db.GetAttributes(processID)
		if err != nil {
			return nil, err
		}

		if len(attributes) == 0 {
			attributes = make([]core.Attribute, 0)
		}

		if len(targetRuntimeIDs) == 1 && targetRuntimeIDs[0] == "*" {
			targetRuntimeIDs = []string{}
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

		processSpec := core.CreateProcessSpec(name, fn, args, targetColonyID, targetRuntimeIDs, runtimeType, maxWaitTime, maxExecTime, maxRetries, env, dependencies, priority)
		process := core.CreateProcessFromDB(processSpec, processID, assignedRuntimeID, isAssigned, state, submissionTime, startTime, endTime, waitDeadline, execDeadline, errs, retries, attributes)
		process.Results = results
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 AND SUBMISSION_TIME BETWEEN NOW() - INTERVAL '1 seconds' * $3 AND NOW() ORDER BY SUBMISSION_TIME DESC`
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

func (db *PQDatabase) FindProcessesByRuntimeID(colonyID string, runtimeID string, seconds int, state int) ([]*core.Process, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND ASSIGNED_RUNTIME_ID=$2 AND STATE=$3 AND SUBMISSION_TIME BETWEEN NOW() - INTERVAL '1 seconds' * $4 AND NOW() ORDER BY SUBMISSION_TIME DESC`
	rows, err := db.postgresql.Query(sqlStatement, colonyID, runtimeID, state, strconv.Itoa(seconds))
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY SUBMISSION_TIME DESC LIMIT $3`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND STATE=$2 ORDER BY START_TIME DESC LIMIT $3`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 ORDER BY START_TIME DESC`
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
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE STATE=$1 ORDER BY START_TIME DESC`
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

func (db *PQDatabase) FindUnassignedProcesses(colonyID string, runtimeID string, runtimeType string, count int, latest bool) ([]*core.Process, error) {
	var sqlStatement string

	// Note: The @> function tests if an array is a subset of another array
	// We need to do that since the TARGET_runtime_IDS can contains many IDs
	if latest {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE RUNTIME_TYPE=$1 AND IS_ASSIGNED=FALSE AND WAIT_FOR_PARENTS=FALSE AND TARGET_COLONY_ID=$2 AND (TARGET_runtime_IDS@>$3 OR TARGET_runtime_IDS@>$4) ORDER BY SUBMISSION_TIME DESC LIMIT $5`
	} else {
		sqlStatement = `SELECT * FROM ` + db.dbPrefix + `PROCESSES WHERE RUNTIME_TYPE=$1 AND IS_ASSIGNED=FALSE AND WAIT_FOR_PARENTS=FALSE AND TARGET_COLONY_ID=$2 AND (TARGET_runtime_IDS@>$3 OR TARGET_runtime_IDS@>$4) ORDER BY SUBMISSION_TIME LIMIT $5`
	}
	rows, err := db.postgresql.Query(sqlStatement, runtimeType, colonyID, pq.Array([]string{runtimeID}), pq.Array([]string{"*"}), count)
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

func (db *PQDatabase) DeleteAllProcessesByColonyID(colonyID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, colonyID)
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
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `PROCESSES WHERE TARGET_COLONY_ID=$1 AND PROCESSGRAPH_ID!=$2`
	_, err := db.postgresql.Exec(sqlStatement, colonyID, "")
	if err != nil {
		return err
	}

	err = db.DeleteAllAttributesInProcessGraphsByColonyID(colonyID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) ResetProcess(process *core.Process) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_RUNTIME_ID=$3, STATE=$4 WHERE PROCESS_ID=$5`
	_, err := db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, process.ID)
	if err != nil {
		return err
	}

	process.SetStartTime(time.Time{})
	process.SetEndTime(time.Time{})
	process.SetAssignedRuntimeID("")
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

func (db *PQDatabase) SetProcessState(processID string, state int) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET STATE=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, state, processID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) SetResults(processID string, results []string) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET RESULTS=$1 WHERE PROCESS_ID=$2`
	_, err := db.postgresql.Exec(sqlStatement, pq.Array(results), processID)
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

func (db *PQDatabase) ResetAllProcesses(process *core.Process) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_RUNTIME_ID=$3, STATE=$4`
	_, err := db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) AssignRuntime(runtimeID string, process *core.Process) error {
	processFromDB, err := db.GetProcessByID(process.ID)
	if err != nil {
		return err
	}

	if processFromDB.IsAssigned {
		return errors.New("Process already assigned")
	}

	startTime := time.Now()
	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=TRUE, START_TIME=$1, ASSIGNED_RUNTIME_ID=$2, STATE=$3 WHERE PROCESS_ID=$4`
	_, err = db.postgresql.Exec(sqlStatement, startTime, runtimeID, core.RUNNING, process.ID)
	if err != nil {
		return err
	}

	process.SetStartTime(startTime)
	process.Assign()
	process.SetAssignedRuntimeID(runtimeID)
	process.SetState(core.RUNNING)

	return nil
}

func (db *PQDatabase) UnassignRuntime(process *core.Process) error {
	endTime := time.Now()

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET IS_ASSIGNED=FALSE, END_TIME=$1, STATE=$2, RETRIES=$3, ASSIGNED_RUNTIME_ID=$4 WHERE PROCESS_ID=$5`
	_, err := db.postgresql.Exec(sqlStatement, endTime, core.WAITING, process.Retries+1, "", process.ID)
	if err != nil {
		return err
	}

	process.SetEndTime(endTime)
	process.Unassign()
	process.SetState(core.WAITING)

	return nil
}

func (db *PQDatabase) MarkSuccessful(process *core.Process) error {
	if process.State == core.FAILED {
		return errors.New("Tried to set failed process as completed")
	}

	if process.State == core.WAITING {
		return errors.New("Tried to set waiting process as completed without being running")
	}

	processFromDB, err := db.GetProcessByID(process.ID)
	if err != nil {
		return err
	}

	if processFromDB.State == core.FAILED {
		return errors.New("Tried to set failed process (from db) as successful")
	}

	if processFromDB.State == core.WAITING {
		return errors.New("Tried to set waiting process (from db) as successful without being running")
	}

	endTime := time.Now()

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET END_TIME=$1, STATE=$2 WHERE PROCESS_ID=$3`
	_, err = db.postgresql.Exec(sqlStatement, endTime, core.SUCCESS, process.ID)
	if err != nil {
		return err
	}

	process.SetEndTime(endTime)
	process.SetState(core.SUCCESS)

	return nil
}

func (db *PQDatabase) MarkFailed(process *core.Process, errs []string) error {
	endTime := time.Now()

	if process.State == core.SUCCESS {
		return errors.New("Tried to set successful process as failed")
	}

	if process.State == core.FAILED {
		return errors.New("Tried to set failed process as failed")
	}

	processFromDB, err := db.GetProcessByID(process.ID)
	if err != nil {
		return err
	}

	if processFromDB.State == core.SUCCESS {
		return errors.New("Tried to set successful (from db) as failed")
	}

	if processFromDB.State == core.FAILED {
		return errors.New("Tried to set failed (from db) as failed")
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `PROCESSES SET END_TIME=$1, STATE=$2 WHERE PROCESS_ID=$3`
	_, err = db.postgresql.Exec(sqlStatement, endTime, core.FAILED, process.ID)
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
