package core

import (
	"colonies/pkg/crypto"
	"time"

	"github.com/google/uuid"
)

const (
	WAITING int = 0
	RUNNING     = 1
	SUCCESS     = 2
	FAILED      = 3
)

type Process struct {
	processID          string
	targetColonyID     string
	targetComputerIDs  []string
	assignedComputerID string
	status             int
	isAssigned         bool
	computerType       string
	submissionTime     time.Time
	startTime          time.Time
	endTime            time.Time
	deadline           time.Time
	timeout            int
	retries            int
	maxRetries         int
	log                string
	mem                int
	cores              int
	gpus               int
}

func CreateProcess(targetColonyID string, targetComputerIDs []string, computerType string, timeout int, maxRetries int, mem int, cores int, gpus int) *Process {
	uuid := uuid.New()
	processID := crypto.GenerateHashFromString(uuid.String()).String()

	process := &Process{processID: processID, targetColonyID: targetColonyID, targetComputerIDs: targetComputerIDs, status: WAITING, isAssigned: false, computerType: computerType, timeout: timeout, maxRetries: maxRetries, mem: mem, cores: cores, gpus: gpus}

	return process
}

func CreateProcessFromDB(processID string, targetColonyID string, targetComputerIDs []string, assignedComputerID string, status int, isAssigned bool, computerType string, submissionTime time.Time, startTime time.Time, endTime time.Time, deadline time.Time, timeout int, retries int, maxRetries int, log string, mem int, cores int, gpus int) *Process {
	return &Process{processID: processID, targetColonyID: targetColonyID, targetComputerIDs: targetComputerIDs, assignedComputerID: assignedComputerID, status: status, isAssigned: isAssigned, computerType: computerType, submissionTime: submissionTime, startTime: startTime, endTime: endTime, deadline: deadline, timeout: timeout, retries: retries, maxRetries: maxRetries, log: log, mem: mem, cores: cores, gpus: gpus}
}

func (process *Process) ID() string {
	return process.processID
}

func (process *Process) TargetColonyID() string {
	return process.targetColonyID
}

func (process *Process) TargetComputerIDs() []string {
	return process.targetComputerIDs
}

func (process *Process) AssignedComputerID() string {
	return process.assignedComputerID
}

func (process *Process) SetAssignedComputerID(computerID string) {
	process.assignedComputerID = computerID
}

func (process *Process) Status() int {
	return process.status
}

func (process *Process) SetStatus(status int) {
	process.status = status
}

func (process *Process) ComputerType() string {
	return process.computerType
}

func (process *Process) SubmissionTime() time.Time {
	return process.submissionTime
}
func (process *Process) SetSubmissionTime(submissionTime time.Time) {
	process.submissionTime = submissionTime
}

func (process *Process) StartTime() time.Time {
	return process.startTime
}

func (process *Process) SetStartTime(startTime time.Time) {
	process.startTime = startTime
}

func (process *Process) EndTime() time.Time {
	return process.endTime
}

func (process *Process) SetEndTime(endTime time.Time) {
	process.endTime = endTime
}

func (process *Process) Deadline() time.Time {
	return process.deadline
}

func (process *Process) Timeout() int {
	return process.timeout
}

func (process *Process) Retries() int {
	return process.retries
}

func (process *Process) MaxRetries() int {
	return process.maxRetries
}

func (process *Process) Log() string {
	return process.log
}

func (process *Process) Mem() int {
	return process.mem
}

func (process *Process) Cores() int {
	return process.cores
}

func (process *Process) GPUs() int {
	return process.gpus
}

func (process *Process) Assigned() bool {
	return process.isAssigned
}

func (process *Process) Assign() {
	process.isAssigned = true
}

func (process *Process) Unassign() {
	process.isAssigned = false
}

func (process *Process) WaitingTime() time.Duration {
	return process.StartTime().Sub(process.SubmissionTime())
}

func (process *Process) ProcessingTime() time.Duration {
	return process.EndTime().Sub(process.StartTime())
}
