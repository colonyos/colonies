package database

import (
	"colonies/pkg/core"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func (db *Database) AddTask(task *core.Task) error {
	targetWorkerIDs := task.TargetWorkerIDs()
	if len(task.TargetWorkerIDs()) == 0 {
		targetWorkerIDs = []string{"*"}
	}

	submissionTime := time.Now()

	sqlStatement := `INSERT INTO  ` + db.dbPrefix + `TASKS (TASK_ID, TARGET_COLONY_ID, TARGET_WORKER_IDS, ASSIGNED_WORKER_ID, STATUS, WORKER_TYPE, SUBMISSION_TIME, START_TIME, END_TIME, DEADLINE, RETRIES, TIMEOUT, MAX_RETRIES, LOG, MEM, CORES, GPUs) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`
	_, err := db.postgresql.Exec(sqlStatement, task.ID(), task.TargetColonyID(), pq.Array(targetWorkerIDs), task.AssignedWorkerID(), core.PENDING, task.WorkerType(), submissionTime, time.Time{}, time.Time{}, task.Deadline(), 0, task.Timeout(), task.MaxRetries(), "", task.Mem(), task.Cores(), task.GPUs())
	if err != nil {
		return err
	}

	task.SetSubmissionTime(submissionTime)

	return nil
}

func (db *Database) parseTasks(rows *sql.Rows) ([]*core.Task, error) {
	var tasks []*core.Task

	for rows.Next() {
		var taskID string
		var targetColonyID string
		var targetWorkerIDs []string
		var assignedWorkerID string
		var status int
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

		if err := rows.Scan(&taskID, &targetColonyID, pq.Array(&targetWorkerIDs), &assignedWorkerID, &status, &workerType, &submissionTime, &startTime, &endTime, &deadline, &timeout, &retries, &maxRetries, &log, &mem, &cores, &gpus); err != nil {
			return nil, err
		}

		task := core.CreateTaskFromDB(taskID, targetColonyID, targetWorkerIDs, assignedWorkerID, status, workerType, submissionTime, startTime, endTime, deadline, timeout, retries, maxRetries, log, mem, cores, gpus)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (db *Database) GetTasks() ([]*core.Task, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `TASKS`
	rows, err := db.postgresql.Query(sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return db.parseTasks(rows)
}

func (db *Database) GetTaskByID(id string) (*core.Task, error) {
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `TASKS WHERE TASK_ID=$1`
	rows, err := db.postgresql.Query(sqlStatement, id)
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

func (db *Database) selectCandidate(candidates []*core.Task) *core.Task {
	if len(candidates) > 0 {
		return candidates[0]
	} else {
		return nil
	}
}

func (db *Database) SearchTask(colonyID string, workerID string) ([]*core.Task, error) {
	var matches []*core.Task

	// Note: The @> function tests if an array is a subset of another array
	// We need to do that since the TARGET_WORKER_IDS can contains many IDs
	sqlStatement := `SELECT * FROM ` + db.dbPrefix + `TASKS WHERE TARGET_COLONY_ID=$1 AND TARGET_WORKER_IDS@>$2 ORDER BY SUBMISSION_TIME LIMIT 1`
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

	sqlStatement = `SELECT * FROM ` + db.dbPrefix + `TASKS WHERE TARGET_COLONY_ID=$1 AND TARGET_WORKER_IDS=$2 ORDER BY SUBMISSION_TIME LIMIT 1`
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

func (db *Database) DeleteTaskByID(taskID string) error {
	sqlStatement := `DELETE FROM ` + db.dbPrefix + `TASKS WHERE TASK_ID=$1`
	_, err := db.postgresql.Exec(sqlStatement, taskID)
	if err != nil {
		return err
	}

	return nil
}
