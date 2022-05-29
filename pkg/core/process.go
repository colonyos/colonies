package core

import (
	"encoding/json"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"

	"github.com/google/uuid"
)

const (
	WAITING int = 0
	RUNNING     = 1
	SUCCESS     = 2
	FAILED      = 3
)

type Process struct {
	ID                string       `json:"processid"`
	AssignedRuntimeID string       `json:"assignedruntimeid"`
	IsAssigned        bool         `json:"isassigned"`
	State             int          `json:"state"`
	SubmissionTime    time.Time    `json:"submissiontime"`
	StartTime         time.Time    `json:"starttime"`
	EndTime           time.Time    `json:"endtime"`
	Deadline          time.Time    `json:"deadline"`
	Retries           int          `json:"retries"`
	Attributes        []*Attribute `json:"attributes"`
	ProcessSpec       *ProcessSpec `json:"spec"`
	WaitingForParents bool         `json:"waitforparents"`
	Parents           []string     `json:"parents"`
	Children          []string     `json:"children"`
	ProcessGraphID    string       `json:"processgraphid"`
}

func CreateProcess(processSpec *ProcessSpec) *Process {
	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	var attributes []*Attribute

	process := &Process{ID: id,
		State:       WAITING,
		IsAssigned:  false,
		Attributes:  attributes,
		ProcessSpec: processSpec,
	}

	return process
}

func CreateProcessFromDB(processSpec *ProcessSpec,
	id string,
	assignedRuntimeID string,
	isAssigned bool,
	state int,
	submissionTime time.Time,
	startTime time.Time,
	endTime time.Time,
	deadline time.Time,
	retries int,
	attributes []*Attribute) *Process {
	return &Process{ID: id,
		AssignedRuntimeID: assignedRuntimeID,
		IsAssigned:        isAssigned,
		State:             state,
		SubmissionTime:    submissionTime,
		StartTime:         startTime,
		EndTime:           endTime,
		Deadline:          deadline,
		Retries:           retries,
		Attributes:        attributes,
		ProcessSpec:       processSpec,
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

func IsProcessArraysEqual(processes1 []*Process, processes2 []*Process) bool {
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
	if process2 == nil {
		return false
	}

	same := true
	if process.ID != process2.ID ||
		process.AssignedRuntimeID != process2.AssignedRuntimeID ||
		process.State != process2.State ||
		process.IsAssigned != process2.IsAssigned ||
		process.SubmissionTime.Unix() != process2.SubmissionTime.Unix() ||
		process.StartTime.Unix() != process2.StartTime.Unix() ||
		process.EndTime.Unix() != process2.EndTime.Unix() ||
		process.Deadline.Unix() != process2.Deadline.Unix() ||
		process.Retries != process2.Retries ||
		process.WaitingForParents != process2.WaitingForParents ||
		process.ProcessGraphID != process2.ProcessGraphID {
		same = false
	}

	if !IsAttributeArraysEqual(process.Attributes, process2.Attributes) {
		same = false
	}

	if !process.ProcessSpec.Equals(process2.ProcessSpec) {
		same = false
	}

	if process.Parents != nil && process2.Parents == nil {
		same = false
	} else if process.Parents == nil && process2.Parents != nil {
		same = false
	} else {
		counter := 0
		for _, parent1 := range process.Parents {
			for _, parent2 := range process.Parents {
				if parent1 == parent2 {
					counter++
				}
			}
		}
		if counter != len(process.Parents) && counter != len(process2.Parents) {
			same = false
		}
	}

	if process.Children != nil && process2.Children == nil {
		same = false
	} else if process.Children == nil && process2.Children != nil {
		same = false
	} else {
		counter := 0
		for _, child1 := range process.Children {
			for _, child2 := range process.Children {
				if child1 == child2 {
					counter++
				}
			}
		}
		if counter != len(process.Children) && counter != len(process2.Children) {
			same = false
		}
	}

	return same
}

func (process *Process) Assign() {
	process.IsAssigned = true
}

func (process *Process) Unassign() {
	process.IsAssigned = false
}

func (process *Process) SetProcessGraphID(processGraphID string) {
	process.ProcessGraphID = processGraphID
}

func (process *Process) SetState(state int) {
	process.State = state
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
	if process.State == WAITING {
		return time.Now().Sub(process.SubmissionTime)
	} else {
		return process.StartTime.Sub(process.SubmissionTime)
	}
}

func (process *Process) ProcessingTime() time.Duration {
	if process.State == RUNNING {
		return time.Now().Sub(process.StartTime)
	} else {
		return process.EndTime.Sub(process.StartTime)
	}
}

func (process *Process) AddParent(parentID string) {
	process.Parents = append(process.Parents, parentID)
}

func (process *Process) AddChild(childID string) {
	process.Children = append(process.Children, childID)
}

func (process *Process) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(process, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
