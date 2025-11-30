package core

import (
	"encoding/json"
	"time"
)

const (
	PENDING      int = 0
	APPROVED         = 1
	REJECTED         = 2
	UNREGISTERED     = 3
)

type Location struct {
	Long        float64 `json:"long"`
	Lat         float64 `json:"lat"`
	Name        string  `json:"name"`
	Description string  `json:"desc"`
}

type GPU struct {
	Name      string `json:"name"`
	Memory    string `json:"mem"`
	Count     int    `json:"count"`
	NodeCount int    `json:"nodecount"`
}

type Software struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type Hardware struct {
	Model        string   `json:"model"`
	Nodes        int      `json:"nodes"`
	CPU          string   `json:"cpu"`
	Cores        int      `json:"cores"`
	Memory       string   `json:"mem"`
	Storage      string   `json:"storage"`
	GPU          GPU      `json:"gpu"`
	Platform     string   `json:"platform"`     // "linux", "darwin", "windows"
	Architecture string   `json:"architecture"` // "amd64", "arm64", "riscv64"
	Network      []string `json:"network"`      // Local network addresses (e.g., "192.168.1.100", "10.0.0.5")
}

type Capabilities struct {
	Hardware []Hardware `json:"hardware"`
	Software []Software `json:"software"`
}

type Project struct {
	AllocatedCPU     int64 `json:"allocatedcpu"`
	UsedCPU          int64 `json:"usedcpu"`
	AllocatedGPU     int64 `json:"allocatedgpu"`
	UsedGPU          int64 `json:"usedgpu"`
	AllocatedStorage int64 `json:"allocatedstorage"`
	UsedStorage      int64 `json:"usedstorage"`
}

type Allocations struct {
	Projects map[string]Project `json:"projects"`
}

type Executor struct {
	ID                string       `json:"executorid"`
	Type              string       `json:"executortype"`
	Name              string       `json:"executorname"`
	ColonyName        string       `json:"colonyname"`
	State             int          `json:"state"`
	RequireFuncReg    bool         `json:"requirefuncreg"`
	CommissionTime    time.Time    `json:"commissiontime"`
	LastHeardFromTime time.Time    `json:"lastheardfromtime"`
	Location          Location     `json:"location"`
	Capabilities      Capabilities `json:"capabilities"`
	Allocations       Allocations  `json:"allocations"`
	BlueprintID       string       `json:"blueprintid,omitempty"`  // Reference to Blueprint (for managed executors)
	BlueprintGen      int64        `json:"blueprintgen,omitempty"` // Blueprint generation this executor belongs to
}

func CreateExecutor(id string,
	executorType string,
	name string,
	colonyName string,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Executor {
	return &Executor{ID: id,
		Type:              executorType,
		Name:              name,
		ColonyName:        colonyName,
		State:             PENDING,
		RequireFuncReg:    false,
		CommissionTime:    commissionTime,
		LastHeardFromTime: lastHeardFromTime,
	}
}

func CreateExecutorFromDB(id string,
	executorType string,
	name string,
	colonyName string,
	state int,
	requireFuncReg bool,
	commissionTime time.Time,
	lastHeardFromTime time.Time) *Executor {

	executor := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
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
	jsonBytes, err := json.Marshal(executors)
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

	if executor.ColonyName != executor2.ColonyName {
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

	if executor.Location.Name != executor2.Location.Name {
		same = false
	}

	if !IsHardwareArraysEqual(executor.Capabilities.Hardware, executor2.Capabilities.Hardware) {
		same = false
	}

	if !IsSoftwareArraysEqual(executor.Capabilities.Software, executor2.Capabilities.Software) {
		same = false
	}

	if executor.Allocations.Projects == nil && executor2.Allocations.Projects != nil {
		same = false
	}

	if executor.Allocations.Projects != nil && executor2.Allocations.Projects == nil {
		same = false
	}

	if executor.Allocations.Projects != nil && executor2.Allocations.Projects != nil {
		if len(executor.Allocations.Projects) != len(executor2.Allocations.Projects) {
			same = false
		}

		if !IsProjectsEqual(executor.Allocations.Projects, executor2.Allocations.Projects) {
			same = false
		}
	}

	return same
}

func IsHardwareEqual(hw1 Hardware, hw2 Hardware) bool {
	if hw1.Model != hw2.Model {
		return false
	}
	if hw1.Nodes != hw2.Nodes {
		return false
	}
	if hw1.CPU != hw2.CPU {
		return false
	}
	if hw1.Cores != hw2.Cores {
		return false
	}
	if hw1.Memory != hw2.Memory {
		return false
	}
	if hw1.Storage != hw2.Storage {
		return false
	}
	if hw1.Platform != hw2.Platform {
		return false
	}
	if hw1.Architecture != hw2.Architecture {
		return false
	}
	if hw1.GPU.Name != hw2.GPU.Name {
		return false
	}
	if hw1.GPU.Memory != hw2.GPU.Memory {
		return false
	}
	if hw1.GPU.Count != hw2.GPU.Count {
		return false
	}
	if hw1.GPU.NodeCount != hw2.GPU.NodeCount {
		return false
	}
	if len(hw1.Network) != len(hw2.Network) {
		return false
	}
	for i, n := range hw1.Network {
		if n != hw2.Network[i] {
			return false
		}
	}
	return true
}

func IsHardwareArraysEqual(hw1 []Hardware, hw2 []Hardware) bool {
	if len(hw1) != len(hw2) {
		return false
	}
	for i, h := range hw1 {
		if !IsHardwareEqual(h, hw2[i]) {
			return false
		}
	}
	return true
}

func IsSoftwareEqual(sw1 Software, sw2 Software) bool {
	if sw1.Name != sw2.Name {
		return false
	}
	if sw1.Type != sw2.Type {
		return false
	}
	if sw1.Version != sw2.Version {
		return false
	}
	return true
}

func IsSoftwareArraysEqual(sw1 []Software, sw2 []Software) bool {
	if len(sw1) != len(sw2) {
		return false
	}
	for i, s := range sw1 {
		if !IsSoftwareEqual(s, sw2[i]) {
			return false
		}
	}
	return true
}

func IsProjectEqual(project1 Project, project2 Project) bool {
	if project1.AllocatedCPU != project2.AllocatedCPU {
		return false
	}

	if project1.UsedCPU != project2.UsedCPU {
		return false
	}

	if project1.AllocatedGPU != project2.AllocatedGPU {
		return false
	}

	if project1.UsedGPU != project2.UsedGPU {
		return false
	}

	if project1.AllocatedStorage != project2.AllocatedStorage {
		return false
	}

	if project1.UsedStorage != project2.UsedStorage {
		return false
	}

	return true
}

func IsProjectsEqual(projects1 map[string]Project, projects2 map[string]Project) bool {
	if len(projects1) != len(projects2) {
		return false
	}

	for key, project1 := range projects1 {
		if !IsProjectEqual(project1, projects2[key]) {
			return false
		}
	}

	return true
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

func (executor *Executor) IsUnregistered() bool {
	if executor.State == UNREGISTERED {
		return true
	}

	return false
}

func (executor *Executor) Unregister() {
	executor.State = UNREGISTERED
}

func (executor *Executor) SetID(id string) {
	executor.ID = id
}

func (executor *Executor) SetColonyName(colonyName string) {
	executor.ColonyName = colonyName
}

func (executor *Executor) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(executor)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
