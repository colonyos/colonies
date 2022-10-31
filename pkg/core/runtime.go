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

type Runtime struct {
	ID                string    `json:"runtimeid"`
	RuntimeType       string    `json:"runtimetype"`
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

func CreateRuntime(id string,
	runtimeType string,
	name string,
	colonyID string,
	cpu string,
	cores int,
	mem int,
	gpu string,
	gpus int,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Runtime {
	return &Runtime{ID: id,
		RuntimeType:       runtimeType,
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

func CreateRuntimeFromDB(id string,
	runtimeType string,
	name string,
	colonyID string,
	cpu string,
	cores int,
	mem int,
	gpu string,
	gpus int,
	state int,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Runtime {
	runtime := CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	runtime.State = state
	return runtime
}

func ConvertJSONToRuntime(jsonString string) (*Runtime, error) {
	var runtime *Runtime
	err := json.Unmarshal([]byte(jsonString), &runtime)
	if err != nil {
		return nil, err
	}

	return runtime, nil
}

func ConvertJSONToRuntimeArray(jsonString string) ([]*Runtime, error) {
	var runtimes []*Runtime
	err := json.Unmarshal([]byte(jsonString), &runtimes)
	if err != nil {
		return runtimes, err
	}

	return runtimes, nil
}

func ConvertRuntimeArrayToJSON(runtimes []*Runtime) (string, error) {
	jsonBytes, err := json.MarshalIndent(runtimes, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func IsRuntimeArraysEqual(runtimes1 []*Runtime, runtimes2 []*Runtime) bool {
	counter := 0
	for _, runtime1 := range runtimes1 {
		for _, runtime2 := range runtimes2 {
			if runtime1.Equals(runtime2) {
				counter++
			}
		}
	}

	if counter == len(runtimes1) && counter == len(runtimes2) {
		return true
	}

	return false
}

func (runtime *Runtime) Equals(runtime2 *Runtime) bool {
	if runtime2 == nil {
		return false
	}

	if runtime.ID == runtime2.ID &&
		runtime.RuntimeType == runtime2.RuntimeType &&
		runtime.Name == runtime2.Name &&
		runtime.ColonyID == runtime2.ColonyID &&
		runtime.CPU == runtime2.CPU &&
		runtime.Cores == runtime2.Cores &&
		runtime.Mem == runtime2.Mem &&
		runtime.GPU == runtime2.GPU &&
		runtime.GPUs == runtime2.GPUs &&
		runtime.State == runtime2.State {
		return true
	}

	return false
}

func (runtime *Runtime) IsApproved() bool {
	if runtime.State == APPROVED {
		return true
	}

	return false
}

func (runtime *Runtime) IsRejected() bool {
	if runtime.State == REJECTED {
		return true
	}

	return false
}

func (runtime *Runtime) IsPending() bool {
	if runtime.State == PENDING {
		return true
	}

	return false
}

func (runtime *Runtime) Approve() {
	runtime.State = APPROVED
}

func (runtime *Runtime) Reject() {
	runtime.State = REJECTED
}

func (runtime *Runtime) SetID(id string) {
	runtime.ID = id
}

func (runtime *Runtime) SetColonyID(colonyID string) {
	runtime.ColonyID = colonyID
}

func (runtime *Runtime) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(runtime, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
