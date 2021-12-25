package core

import (
	"colonies/pkg/crypto"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	WAITING int = 0
	RUNNING     = 1
	SUCCESS     = 2
	FAILED      = 3
)

// TODO: This code should be refactored so that it contains a ProcessSpec instead of all this redundant information

type Process struct {
	ID                string       `json:"processid"`
	TargetColonyID    string       `json:"targetcolonyid"`
	TargetRuntimeIDs  []string     `json:"targetruntimeids"`
	AssignedRuntimeID string       `json:"assignedruntimeid"`
	Status            int          `json:"status"`
	IsAssigned        bool         `json:"isassigned"`
	RuntimeType       string       `json:"runtimetype"`
	SubmissionTime    time.Time    `json:"submissiontime"`
	StartTime         time.Time    `json:"starttime"`
	EndTime           time.Time    `json:"endtime"`
	Deadline          time.Time    `json:"deadline"`
	Timeout           int          `json:"timeout"`
	Retries           int          `json:"retries"`
	MaxRetries        int          `json:"maxretries"`
	Mem               int          `json:"mem"`
	Cores             int          `json:"cores"`
	GPUs              int          `json:"gpus"`
	Attributes        []*Attribute `json:"attributes"`
}

func CreateProcess(targetColonyID string, targetRuntimeIDs []string, runtimeType string, timeout int, maxRetries int, mem int, cores int, gpus int) *Process {
	uuid := uuid.New()
	id := crypto.GenerateHashFromString(uuid.String()).String()

	var attributes []*Attribute

	process := &Process{ID: id,
		TargetColonyID:   targetColonyID,
		TargetRuntimeIDs: targetRuntimeIDs,
		Status:           WAITING,
		IsAssigned:       false,
		RuntimeType:      runtimeType,
		Timeout:          timeout,
		MaxRetries:       maxRetries,
		Mem:              mem,
		Cores:            cores,
		GPUs:             gpus,
		Attributes:       attributes,
	}

	return process
}

func CreateProcessFromDB(id string, targetColonyID string, targetRuntimeIDs []string, assignedRuntimeID string, status int, isAssigned bool, runtimeType string, submissionTime time.Time, startTime time.Time, endTime time.Time, deadline time.Time, timeout int, retries int, maxRetries int, mem int, cores int, gpus int, attributes []*Attribute) *Process {
	return &Process{ID: id,
		TargetColonyID:    targetColonyID,
		TargetRuntimeIDs:  targetRuntimeIDs,
		AssignedRuntimeID: assignedRuntimeID,
		Status:            status,
		IsAssigned:        isAssigned,
		RuntimeType:       runtimeType,
		SubmissionTime:    submissionTime,
		StartTime:         startTime,
		EndTime:           endTime,
		Deadline:          deadline,
		Timeout:           timeout,
		Retries:           retries,
		MaxRetries:        maxRetries,
		Mem:               mem,
		Cores:             cores,
		GPUs:              gpus,
		Attributes:        attributes,
	}
}

func ConvertJSONToProcess(jsonString string) (*Process, error) {
	var process *Process
	err := json.Unmarshal([]byte(jsonString), &process)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func ConvertProcessArrayToJSON(processes []*Process) (string, error) {
	jsonBytes, err := json.MarshalIndent(processes, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToProcessArray(jsonString string) ([]*Process, error) {
	var processes []*Process
	err := json.Unmarshal([]byte(jsonString), &processes)
	if err != nil {
		return processes, err
	}

	return processes, nil
}

func IsProcessArrayEqual(processes1 []*Process, processes2 []*Process) bool {
	counter := 0
	for _, process1 := range processes1 {
		for _, process2 := range processes2 {
			if process1.Equals(process2) {
				counter++
			}
		}
	}

	if counter == len(processes1) && counter == len(processes2) {
		return true
	}

	return false
}

func (process *Process) Equals(process2 *Process) bool {
	same := true
	if process.ID != process2.ID &&
		process.TargetColonyID != process2.TargetColonyID &&
		process.AssignedRuntimeID != process2.AssignedRuntimeID &&
		process.Status != process2.Status &&
		process.IsAssigned != process2.IsAssigned &&
		process.RuntimeType != process2.RuntimeType &&
		process.SubmissionTime != process2.SubmissionTime &&
		process.StartTime != process2.StartTime &&
		process.EndTime != process2.EndTime &&
		process.Deadline != process2.Deadline &&
		process.Timeout != process2.Timeout &&
		process.Retries != process2.Retries &&
		process.MaxRetries != process2.MaxRetries &&
		process.Mem != process2.Mem &&
		process.Cores != process2.Cores &&
		process.GPUs != process2.GPUs {
		same = false
	}

	if process.TargetRuntimeIDs != nil && process2.TargetRuntimeIDs == nil {
		same = false
	} else if process.TargetRuntimeIDs == nil && process2.TargetRuntimeIDs != nil {
		same = false
	} else {
		counter := 0
		for _, targetRuntimeID1 := range process.TargetRuntimeIDs {
			for _, targetRuntimeID2 := range process2.TargetRuntimeIDs {
				if targetRuntimeID1 == targetRuntimeID2 {
					counter++
				}
			}
		}
		if counter != len(process.TargetRuntimeIDs) && counter != len(process2.TargetRuntimeIDs) {
			same = false
		}
	}

	if !IsAttributeArraysEqual(process.Attributes, process2.Attributes) {
		same = false
	}

	return same
}

func (process *Process) Assign() {
	process.IsAssigned = true
}

func (process *Process) Unassign() {
	process.IsAssigned = false
}

func (process *Process) SetStatus(status int) {
	process.Status = status
}

func (process *Process) SetAssignedRuntimeID(runtimeID string) {
	process.AssignedRuntimeID = runtimeID
	process.IsAssigned = true
}

func (process *Process) SetAttributes(attributes []*Attribute) {
	process.Attributes = attributes
}

func (process *Process) SetSubmissionTime(submissionTime time.Time) {
	process.SubmissionTime = submissionTime
}

func (process *Process) SetStartTime(startTime time.Time) {
	process.StartTime = startTime
}

func (process *Process) SetEndTime(endTime time.Time) {
	process.EndTime = endTime
}

func (process *Process) WaitingTime() time.Duration {
	return process.StartTime.Sub(process.SubmissionTime)
}

func (process *Process) ProcessingTime() time.Duration {
	return process.EndTime.Sub(process.StartTime)
}

func (process *Process) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(process, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
