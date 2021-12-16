package postgresql

import (
	"colonies/pkg/core"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *PQDatabase) AddTask(task *core.Task) error {
	targetWorkerIDs := task.TargetWorkerIDs()
	if len(task.TargetWorkerIDs()) == 0 {
		targetWorkerIDs = []string{"*"}
	}

	submissionTime := time.Now()

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `TASKS (TASK_ID, TARGET_COLONY_ID, TARGET_WORKER_IDS, ASSIGNED_WORKER_ID, STATUS, IS_ASSIGNED, WORKER_TYPE, SUBMISSION_TIME, START_TIME, END_TIME, DEADLINE, RETRIES, TIMEOUT, MAX_RETRIES, LOG, MEM, CORES, GPUs) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`
	_, err := db.postgresql.Exec(sqlStatement, task.ID(), task.TargetColonyID(), pq.Array(targetWorkerIDs), task.AssignedWorkerID(), task.Status(), task.Assigned(), task.WorkerType(), submissionTime, time.Time{}, time.Time{}, task.Deadline(), 0, task.Timeout(), task.MaxRetries(), "", task.Mem(), task.Cores(), task.GPUs())
	if err != nil {
		return err
	}

	task.SetSubmissionTime(submissionTime)

	return nil
}

func (db *PQDatabase) parseTasks(rows *sql.Rows) ([]*core.Task, error) {
	var tasks []*core.Task

	for rows.Next() {
		var taskID string
		var targetColonyID string
		var targetWorkerIDs []string
		var assignedWorkerID string
		var status int
		var isAssigned bool
		var workerType string
		var submissionTime time.Time
		var startTime time.Time
		var endTime time.Time
		var deadline time.Time
		var timeout int
		var retries int
		var maxRetries int
		var log string
		var mem int
		var cores int
		var gpus int

		if err := rows.Scan(&taskID, &targetColonyID, pq.Array(&targetWorkerIDs), &assignedWorkerID, &status, &isAssigned, &workerType, &submissionTime, &startTime, &endTime, &deadline, &timeout, &retries, &maxRetries, &log, &mem, &cores, &gpus); err != nil {
			return nil, err
		}

		task := core.CreateTaskFromDB(taskID, targetColonyID, targetWorkerIDs, assignedWorkerID, status, isAssigned, workerType, submissionTime, startTime, endTime, deadline, timeout, retries, maxRetries, log, mem, cores, gpus)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (db *PQDatabase) GetTasks() ([]*core.Task, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `TASKS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseTasks(rows)
}

func (db *PQDatabase) GetTaskByID(taskID string) (*core.Task, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `TASKS WHERE TASK_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, taskID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks, err := db.parseTasks(rows)
	if err != nil {
		return nil, err
	}

	if len(tasks) > 1 {
		return nil, errors.New("expected one task, task id should be unique")
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	return tasks[0], nil
}

func (db *PQDatabase) selectCandidate(candidates []*core.Task) *core.Task {
	if len(candidates) > 0 {
		return candidates[0]
	} else {
		return nil
	}
}

func (db *PQDatabase) SearchTasks(colonyID string, workerID string) ([]*core.Task, error) {
	var matches []*core.Task

	// Note: The @> function tests if an array is a subset of another array
	// We need to do that since the TARGET_WORKER_IDS can contains many IDs
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `TASKS WHERE IS_ASSIGNED=FALSE AND TARGET_COLONY_ID=$1 AND TARGET_WORKER_IDS@>$2 ORDER BY SUBMISSION_TIME LIMIT 1`
	rows1, err := db.postgresql.Query(sqlStatement, colonyID, pq.Array([]string{workerID}))
	if err != nil {
		return nil, err
	}

	defer rows1.Close()

	tasks, err := db.parseTasks(rows1)
	if err != nil {
		return nil, err
	}

	if len(tasks) > 0 {
		matches = append(matches, tasks[0])
	}

	sqlStatement = `SELECT * FROM ` + db.dbPrefix + `TASKS WHERE IS_ASSIGNED=FALSE AND TARGET_COLONY_ID=$1 AND TARGET_WORKER_IDS=$2 ORDER BY SUBMISSION_TIME LIMIT 1`
	rows2, err := db.postgresql.Query(sqlStatement, colonyID, pq.Array([]string{"*"}))
	if err != nil {
		return nil, err
	}

	defer rows2.Close()

	tasks, err = db.parseTasks(rows2)
	if err != nil {
		return nil, err
	}

	if len(tasks) > 0 {
		matches = append(matches, tasks[0])
	}

	return matches, nil
}

func (db *PQDatabase) DeleteTaskByID(taskID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `TASKS WHERE TASK_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, taskID)
	if err != nil {
		return err
	}

	// TODO test this code
	err = db.DeleteAllAttributesByTaskID(taskID)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) DeleteAllTasks() error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `TASKS`
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

func (db *PQDatabase) ResetTask(task *core.Task) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `TASKS SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_WORKER_ID=$3, STATUS=$4 WHERE TASK_ID=$5`
	_, err := db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING, task.ID())
	if err != nil {
		return err
	}

	task.SetStartTime(time.Time{})
	task.SetEndTime(time.Time{})
	task.SetAssignedWorkerID("")
	task.SetStatus(core.WAITING)

	return nil
}

func (db *PQDatabase) ResetAllTasks(task *core.Task) error {
	sqlStatement := `UPDATE ` + db.dbPrefix + `TASKS SET IS_ASSIGNED=FALSE, START_TIME=$1, END_TIME=$2, ASSIGNED_WORKER_ID=$3, STATUS=$4`
	_, err := db.postgresql.Exec(sqlStatement, time.Time{}, time.Time{}, "", core.WAITING)
	if err != nil {
		return err
	}

	return nil
}

func (db *PQDatabase) AssignWorker(workerID string, task *core.Task) error {
	startTime := time.Now()

	sqlStatement := `UPDATE ` + db.dbPrefix + `TASKS SET IS_ASSIGNED=TRUE, START_TIME=$1, ASSIGNED_WORKER_ID=$2, STATUS=$3 WHERE TASK_ID=$4`
	_, err := db.postgresql.Exec(sqlStatement, startTime, workerID, core.RUNNING, task.ID())
	if err != nil {
		return err
	}

	task.SetStartTime(startTime)
	task.Assign()
	task.SetAssignedWorkerID(workerID)
	task.SetStatus(core.RUNNING)

	return nil
}

func (db *PQDatabase) UnassignWorker(task *core.Task) error {
	endTime := time.Now()

	sqlStatement := `UPDATE ` + db.dbPrefix + `TASKS SET IS_ASSIGNED=FALSE, END_TIME=$1, STATUS=$2 WHERE TASK_ID=$3`
	_, err := db.postgresql.Exec(sqlStatement, endTime, core.FAILED, task.ID())
	if err != nil {
		return err
	}

	task.SetEndTime(endTime)
	task.Unassign()
	task.SetStatus(core.FAILED)

	return nil
}

func (db *PQDatabase) MarkSuccessful(task *core.Task) error {
	if task.Status() == core.FAILED {
		return errors.New("tried to set failed task as completed")
	}

	if task.Status() == core.WAITING {
		return errors.New("tried to set waiting task as completed without being running")
	}

	taskFromDB, err := db.GetTaskByID(task.ID())
	if err != nil {
		return err
	}

	if taskFromDB.Status() == core.FAILED {
		return errors.New("tried to set failed task (from db) as successful")
	}

	if taskFromDB.Status() == core.WAITING {
		return errors.New("tried to set waiting task (from db) as successful without being running")
	}

	endTime := time.Now()

	sqlStatement := `UPDATE ` + db.dbPrefix + `TASKS SET END_TIME=$1, STATUS=$2 WHERE TASK_ID=$3`
	_, err = db.postgresql.Exec(sqlStatement, endTime, core.SUCCESS, task.ID())
	if err != nil {
		return err
	}

	task.SetEndTime(endTime)
	task.SetStatus(core.SUCCESS)

	return nil
}

func (db *PQDatabase) MarkFailed(task *core.Task) error {
	endTime := time.Now()

	if task.Status() == core.SUCCESS {
		return errors.New("tried to set successful task as failed")
	}

	if task.Status() == core.WAITING {
		return errors.New("tried to set waiting task as failed without being running")
	}

	taskFromDB, err := db.GetTaskByID(task.ID())
	if err != nil {
		return err
	}

	if taskFromDB.Status() == core.SUCCESS {
		return errors.New("tried to set successful (from db) as failed")
	}

	if taskFromDB.Status() == core.WAITING {
		return errors.New("tried to set successful task (from db) as failed without being running")
	}

	sqlStatement := `UPDATE ` + db.dbPrefix + `TASKS SET END_TIME=$1, STATUS=$2 WHERE TASK_ID=$3`
	_, err = db.postgresql.Exec(sqlStatement, endTime, core.FAILED, task.ID())
	if err != nil {
		return err
	}

	task.SetEndTime(endTime)
	task.SetStatus(core.SUCCESS)

	return nil
}

func (db *PQDatabase) NumberOfTasks() (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `TASKS`
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

func (db *PQDatabase) countTasks(status int) (int, error) {
	sqlStatement := `SELECT COUNT(*) FROM ` + db.dbPrefix + `TASKS WHERE STATUS=$1`
	rows, err := db.postgresql.Query(sqlStatement, status)
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

func (db *PQDatabase) NumberOfRunningTasks() (int, error) {
	return db.countTasks(core.RUNNING)
}

func (db *PQDatabase) NumberOfSuccessfulTasks() (int, error) {
	return db.countTasks(core.SUCCESS)
}

func (db *PQDatabase) NumberOfFailedTasks() (int, error) {
	return db.countTasks(core.FAILED)
}
