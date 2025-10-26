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

const NOTSET = -1

type Process struct {
	ID                 string        `json:"processid"`
	InitiatorID        string        `json:"initiatorid"`
	InitiatorName      string        `json:"initiatorname"`
	AssignedExecutorID string        `json:"assignedexecutorid"`
	IsAssigned         bool          `json:"isassigned"`
	State              int           `json:"state"`
	PriorityTime       int64         `json:"prioritytime"`
	SubmissionTime     time.Time     `json:"submissiontime"`
	StartTime          time.Time     `json:"starttime"`
	EndTime            time.Time     `json:"endtime"`
	WaitDeadline       time.Time     `json:"waitdeadline"`
	ExecDeadline       time.Time     `json:"execdeadline"`
	Retries            int           `json:"retries"`
	Attributes         []Attribute   `json:"attributes"`
	FunctionSpec       FunctionSpec  `json:"spec"`
	WaitForParents     bool          `json:"waitforparents"`
	Parents            []string      `json:"parents"`
	Children           []string      `json:"children"`
	ProcessGraphID     string        `json:"processgraphid"`
	Input              []interface{} `json:"in"`
	Output             []interface{} `json:"out"`
	Errors             []string      `json:"errors"`
}

func CreateProcess(funcSpec *FunctionSpec) *Process {
	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	var attributes []Attribute

	process := &Process{ID: id,
		State:        WAITING,
		IsAssigned:   false,
		Attributes:   attributes,
		FunctionSpec: *funcSpec,
		Input:        make([]interface{}, 0),
		Output:       make([]interface{}, 0),
		Errors:       make([]string, 0),
	}

	return process
}

func CreateProcessFromDB(funcSpec *FunctionSpec,
	id string,
	assignedExecutorID string,
	isAssigned bool,
	state int,
	priorityTime int64,
	submissionTime time.Time,
	startTime time.Time,
	endTime time.Time,
	waitDeadline time.Time,
	execDeadline time.Time,
	errors []string,
	retries int,
	attributes []Attribute) *Process {
	return &Process{ID: id,
		AssignedExecutorID: assignedExecutorID,
		IsAssigned:         isAssigned,
		State:              state,
		PriorityTime:       priorityTime,
		SubmissionTime:     submissionTime,
		StartTime:          startTime,
		EndTime:            endTime,
		WaitDeadline:       waitDeadline,
		ExecDeadline:       execDeadline,
		Errors:             errors,
		Retries:            retries,
		Attributes:         attributes,
		FunctionSpec:       *funcSpec,
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
	jsonBytes, err := json.MarshalIndent(processes, "", "  ")
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
		process.InitiatorID != process2.InitiatorID ||
		process.InitiatorName != process2.InitiatorName ||
		process.AssignedExecutorID != process2.AssignedExecutorID ||
		process.State != process2.State ||
		process.PriorityTime != process2.PriorityTime ||
		process.IsAssigned != process2.IsAssigned ||
		process.SubmissionTime.Unix() != process2.SubmissionTime.Unix() ||
		process.StartTime.Unix() != process2.StartTime.Unix() ||
		process.EndTime.Unix() != process2.EndTime.Unix() ||
		process.WaitDeadline.Unix() != process2.WaitDeadline.Unix() ||
		process.ExecDeadline.Unix() != process2.ExecDeadline.Unix() ||
		process.Retries != process2.Retries ||
		process.WaitForParents != process2.WaitForParents ||
		process.ProcessGraphID != process2.ProcessGraphID {
		same = false
	}

	if !same {
		return false
	}

	counter := 0
	for _, r1 := range process.Output {
		for _, r2 := range process2.Output {
			if r1 == r2 {
				counter++
			}
		}
	}
	if counter != len(process.Output) && counter != len(process2.Output) {
		same = false
	}

	if !same {
		return false
	}

	counter = 0
	for _, r1 := range process.Input {
		for _, r2 := range process2.Input {
			if r1 == r2 {
				counter++
			}
		}
	}
	if counter != len(process.Input) && counter != len(process2.Input) {
		same = false
	}

	if !same {
		return false
	}

	if !IsAttributeArraysEqual(process.Attributes, process2.Attributes) {
		same = false
	}

	if !same {
		return false
	}

	if !process.FunctionSpec.Equals(&process2.FunctionSpec) {
		same = false
	}

	if !same {
		return false
	}

	counter = 0
	for _, parent1 := range process.Parents {
		for _, parent2 := range process2.Parents {
			if parent1 == parent2 {
				counter++
			}
		}
	}
	if counter != len(process.Parents) && counter != len(process2.Parents) {
		same = false
	}

	if !same {
		return false
	}

	counter = 0
	for _, child1 := range process.Children {
		for _, child2 := range process2.Children {
			if child1 == child2 {
				counter++
			}
		}
	}
	if counter != len(process.Children) && counter != len(process2.Children) {
		same = false
	}

	if !same {
		return false
	}

	counter = 0
	for _, r1 := range process.Output {
		for _, r2 := range process2.Output {
			if r1 == r2 {
				counter++
			}
		}
	}
	if counter != len(process.Output) && counter != len(process2.Output) {
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

func (process *Process) SetProcessGraphID(processGraphID string) {
	process.ProcessGraphID = processGraphID
}

func (process *Process) SetState(state int) {
	process.State = state
}

func (process *Process) SetAssignedExecutorID(executorID string) {
	process.AssignedExecutorID = executorID
	process.IsAssigned = true
}

func (process *Process) SetAttributes(attributes []Attribute) {
	process.Attributes = attributes
}

// -50000 - 50000
func (process *Process) SetSubmissionTime(submissionTime time.Time) {
	process.SubmissionTime = submissionTime
	var dt int64
	dt = -1000000000 * 60 * 60 * 24
	process.PriorityTime = int64(process.FunctionSpec.Priority)*dt + submissionTime.UnixNano()
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
	jsonBytes, err := json.MarshalIndent(process, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (process *Process) Clone() *Process {
	processCopy := new(Process)
	*processCopy = *process
	return processCopy
}
