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
	Long float64 `json:"long"`
	Lat  float64 `json:"lat"`
}

type Executor struct {
	ID                string    `json:"id"`
	Type       string    `json:"type"`
	Name              string    `json:"name"`
	ColonyID          string    `json:"colonyid"`
	CPU               string    `json:"cpu"`
	Cores             int       `json:"cores"`
	Mem               int       `json:"mem"`
	GPU               string    `json:"gpu"`
	GPUs              int       `json:"gpus"`
	State             int       `json:"state"`
	CommissionTime    time.Time `json:"commissiontime"`
	LastHeardFromTime time.Time `json:"lastheardfromtime"`
	Location          Location  `json:"location"`
}

func CreateExecutor(id string,
	executorType string,
	name string,
	colonyID string,
	cpu string,
	cores int,
	mem int,
	gpu string,
	gpus int,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Executor {
	return &Executor{ID: id,
		Type:       executorType,
		Name:              name,
		ColonyID:          colonyID,
		CPU:               cpu,
		Cores:             cores,
		Mem:               mem,
		GPU:               gpu,
		GPUs:              gpus,
		State:             PENDING,
		CommissionTime:    commissionTime,
		LastHeardFromTime: lastHeardFromTime}
}

func CreateExecutorFromDB(id string,
	executorType string,
	name string,
	colonyID string,
	cpu string,
	cores int,
	mem int,
	gpu string,
	gpus int,
	state int,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Executor {
	executor := CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	executor.State = state
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

	if executor.ID == executor2.ID &&
		executor.Type == executor2.Type &&
		executor.Name == executor2.Name &&
		executor.ColonyID == executor2.ColonyID &&
		executor.CPU == executor2.CPU &&
		executor.Cores == executor2.Cores &&
		executor.Mem == executor2.Mem &&
		executor.GPU == executor2.GPU &&
		executor.GPUs == executor2.GPUs &&
		executor.State == executor2.State {
		return true
	}

	return false
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
