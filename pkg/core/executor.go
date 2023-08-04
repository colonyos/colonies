package core

import (
	"encoding/json"
	"time"
)

const (
	PENDING  int = 0
	APPROVED     = 1
	REJECTED     = 2
)

type Location struct {
	Long        float64 `json:"long"`
	Lat         float64 `json:"lat"`
	Description string  `json:"desc"`
}

type GPU struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Software struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type Hardware struct {
	Model   string `json:"model"`
	CPU     string `json:"cpu"`
	Memory  string `json:"mem"`
	Storage string `json:"storage"`
	GPU     GPU    `json:"gpu"`
}

type Capabilities struct {
	Hardware Hardware `json:"hardware"`
	Software Software `json:"software"`
}

type Executor struct {
	ID                string       `json:"executorid"`
	Type              string       `json:"executortype"`
	Name              string       `json:"executorname"`
	ColonyID          string       `json:"colonyid"`
	State             int          `json:"state"`
	RequireFuncReg    bool         `json:"requirefuncreg"`
	CommissionTime    time.Time    `json:"commissiontime"`
	LastHeardFromTime time.Time    `json:"lastheardfromtime"`
	Location          Location     `json:"location"`
	Capabilities      Capabilities `json:"capbilities"`
}

func CreateExecutor(id string,
	executorType string,
	name string,
	colonyID string,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Executor {
	return &Executor{ID: id,
		Type:              executorType,
		Name:              name,
		ColonyID:          colonyID,
		State:             PENDING,
		RequireFuncReg:    false,
		CommissionTime:    commissionTime,
		LastHeardFromTime: lastHeardFromTime,
	}
}

func CreateExecutorFromDB(id string,
	executorType string,
	name string,
	colonyID string,
	state int,
	requireFuncReg bool,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Executor {
	executor := CreateExecutor(id, executorType, name, colonyID, commissionTime, lastHeardFromTime)
	executor.State = state
	executor.RequireFuncReg = requireFuncReg
	return executor
}

func ConvertJSONToExecutor(jsonString string) (*Executor, error) {
	var executor *Executor
	err := json.Unmarshal([]byte(jsonString), &executor)
	if err != nil {
		return nil, err
	}

	return executor, nil
}

func ConvertJSONToExecutorArray(jsonString string) ([]*Executor, error) {
	var executors []*Executor
	err := json.Unmarshal([]byte(jsonString), &executors)
	if err != nil {
		return executors, err
	}

	return executors, nil
}

func ConvertExecutorArrayToJSON(executors []*Executor) (string, error) {
	jsonBytes, err := json.MarshalIndent(executors, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func IsExecutorArraysEqual(executors1 []*Executor, executors2 []*Executor) bool {
	counter := 0
	for _, executor1 := range executors1 {
		for _, executor2 := range executors2 {
			if executor1.Equals(executor2) {
				counter++
			}
		}
	}

	if counter == len(executors1) && counter == len(executors2) {
		return true
	}

	return false
}

func (executor *Executor) Equals(executor2 *Executor) bool {
	if executor2 == nil {
		return false
	}

	same := true
	if executor.ID != executor2.ID {
		same = false
	}

	if executor.Type != executor2.Type {
		same = false
	}

	if executor.Name != executor2.Name {
		same = false
	}

	if executor.ColonyID != executor2.ColonyID {
		same = false
	}

	if executor.State != executor2.State {
		same = false
	}

	if executor.RequireFuncReg != executor2.RequireFuncReg {
		same = false
	}

	if executor.Location.Lat != executor2.Location.Lat {
		same = false
	}

	if executor.Location.Long != executor2.Location.Long {
		same = false
	}

	if executor.Location.Description != executor2.Location.Description {
		same = false
	}

	if executor.Capabilities.Hardware.CPU != executor2.Capabilities.Hardware.CPU {
		same = false
	}

	if executor.Capabilities.Hardware.Memory != executor2.Capabilities.Hardware.Memory {
		same = false
	}

	if executor.Capabilities.Hardware.Storage != executor2.Capabilities.Hardware.Storage {
		same = false
	}

	if executor.Capabilities.Hardware.GPU.Name != executor2.Capabilities.Hardware.GPU.Name {
		same = false
	}

	if executor.Capabilities.Hardware.GPU.Count != executor2.Capabilities.Hardware.GPU.Count {
		same = false
	}

	if executor.Capabilities.Software.Name != executor2.Capabilities.Software.Name {
		same = false
	}

	if executor.Capabilities.Software.Type != executor2.Capabilities.Software.Type {
		same = false
	}

	if executor.Capabilities.Software.Version != executor2.Capabilities.Software.Version {
		same = false
	}

	return same
}

func (executor *Executor) IsApproved() bool {
	if executor.State == APPROVED {
		return true
	}

	return false
}

func (executor *Executor) IsRejected() bool {
	if executor.State == REJECTED {
		return true
	}

	return false
}

func (executor *Executor) IsPending() bool {
	if executor.State == PENDING {
		return true
	}

	return false
}

func (executor *Executor) Approve() {
	executor.State = APPROVED
}

func (executor *Executor) Reject() {
	executor.State = REJECTED
}

func (executor *Executor) SetID(id string) {
	executor.ID = id
}

func (executor *Executor) SetColonyID(colonyID string) {
	executor.ColonyID = colonyID
}

func (executor *Executor) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(executor, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
