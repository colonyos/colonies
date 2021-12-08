package core

import (
	"colonies/pkg/crypto"
	"time"

	"github.com/google/uuid"
)

const (
	WAITING  int = 0
	RUNNING      = 1
	SUCCESS      = 2
	FAILED       = 3
	ASSIGNED     = 4
)

type Task struct {
	taskID           string
	targetColonyID   string
	targetWorkerIDs  []string
	assignedWorkerID string
	status           int
	workerType       string
	submissionTime   time.Time
	startTime        time.Time
	endTime          time.Time
	deadline         time.Time
	timeout          int
	retries          int
	maxRetries       int
	log              string
	mem              int
	cores            int
	gpus             int
}

func CreateTask(targetColonyID string, targetWorkerIDs []string, workerType string, timeout int, maxRetries int, mem int, cores int, gpus int) *Task {
	uuid := uuid.New()
	taskID := crypto.GenerateHashFromString(uuid.String()).String()

	task := &Task{taskID: taskID, targetColonyID: targetColonyID, targetWorkerIDs: targetWorkerIDs, status: WAITING, workerType: workerType, timeout: timeout, maxRetries: maxRetries, mem: mem, cores: cores, gpus: gpus}

	return task
}

func CreateTaskFromDB(taskID string, targetColonyID string, targetWorkerIDs []string, assignedWorkerID string, status int, workerType string, submissionTime time.Time, startTime time.Time, endTime time.Time, deadline time.Time, timeout int, retries int, maxRetries int, log string, mem int, cores int, gpus int) *Task {
	return &Task{taskID: taskID, targetColonyID: targetColonyID, targetWorkerIDs: targetWorkerIDs, assignedWorkerID: assignedWorkerID, status: status, workerType: workerType, submissionTime: submissionTime, startTime: startTime, endTime: endTime, deadline: deadline, timeout: timeout, retries: retries, maxRetries: maxRetries, log: log, mem: mem, cores: cores, gpus: gpus}
}

func (task *Task) ID() string {
	return task.taskID
}

func (task *Task) TargetColonyID() string {
	return task.targetColonyID
}

func (task *Task) TargetWorkerIDs() []string {
	return task.targetWorkerIDs
}

func (task *Task) AssignedWorkerID() string {
	return task.assignedWorkerID
}

func (task *Task) Status() int {
	return task.status
}

func (task *Task) WorkerType() string {
	return task.workerType
}

func (task *Task) SubmissionTime() time.Time {
	return task.submissionTime
}

func (task *Task) StartTime() time.Time {
	return task.startTime
}

func (task *Task) EndTime() time.Time {
	return task.startTime
}

func (task *Task) Deadline() time.Time {
	return task.deadline
}

func (task *Task) Timeout() int {
	return task.timeout
}

func (task *Task) Retries() int {
	return task.retries
}

func (task *Task) MaxRetries() int {
	return task.maxRetries
}

func (task *Task) Log() string {
	return task.log
}

func (task *Task) Mem() int {
	return task.mem
}

func (task *Task) Cores() int {
	return task.cores
}

func (task *Task) GPUs() int {
	return task.gpus
}

func (task *Task) SetSubmissionTime(submissionTime time.Time) {
	task.submissionTime = submissionTime
}
