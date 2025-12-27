package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateExecutor(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := ""
	colonyName := "test_colony_name"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)

	assert.Equal(t, PENDING, executor.State)
	assert.True(t, executor.IsPending())
	assert.False(t, executor.IsApproved())
	assert.False(t, executor.IsRejected())
	assert.Equal(t, id, executor.ID)
	assert.Equal(t, executorType, executor.Type)
	assert.Equal(t, name, executor.Name)
	assert.Equal(t, colonyName, executor.ColonyName)
}

func TestCreateExecutor2(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := ""
	colonyName := "test_colony_name"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor1GPU := GPU{Name: "test_name1", Count: 1, Memory: "11G", NodeCount: 1}
	executor1HW := Hardware{Model: "test_model", CPU: "test_cpu", Cores: 8, Memory: "test_mem", Storage: "test_storage", GPU: executor1GPU, Nodes: 1}
	executor1SW := Software{Name: "test_name1", Type: "test_type1", Version: "test_version1"}
	executor1CAP := Capabilities{Hardware: []Hardware{executor1HW}, Software: []Software{executor1SW}}
	executor1.LocationName = "test_location"
	executor1.Capabilities = executor1CAP

	executor2 := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor2GPU := GPU{Name: "test_name1", Count: 1, Memory: "11G", NodeCount: 1}
	executor2HW := Hardware{Model: "test_model", CPU: "test_cpu", Cores: 8, Memory: "test_mem", Storage: "test_storage", GPU: executor2GPU, Nodes: 1}
	executor2SW := Software{Name: "test_name1", Type: "test_type1", Version: "test_version1"}
	executor2CAP := Capabilities{Hardware: []Hardware{executor2HW}, Software: []Software{executor2SW}}
	executor2.LocationName = "test_location"
	executor2.Capabilities = executor2CAP

	assert.True(t, executor1.Equals(executor2))
	executor2.LocationName = "changed_location"
	assert.False(t, executor1.Equals(executor2))
}

func TestHardwareCoresEqual(t *testing.T) {
	hw1 := Hardware{Model: "test", CPU: "test_cpu", Cores: 8, Memory: "16GB"}
	hw2 := Hardware{Model: "test", CPU: "test_cpu", Cores: 8, Memory: "16GB"}
	hw3 := Hardware{Model: "test", CPU: "test_cpu", Cores: 16, Memory: "16GB"}

	assert.True(t, IsHardwareEqual(hw1, hw2))
	assert.False(t, IsHardwareEqual(hw1, hw3))
}

func TestSetExecutorID(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyName := "test_colony_name"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor.SetID("test_executor_id_set")

	assert.Equal(t, executor.ID, "test_executor_id_set")
}

func TestSetColonyNameOnRimtime(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyName := "test_colony_name"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor.SetColonyName("test_colonyid_set")

	assert.Equal(t, executor.ColonyName, "test_colonyid_set")
}

func TestExecutorEquals(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyName := "test_colony_name"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	assert.True(t, executor1.Equals(executor1))

	executorWithAlloc := CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	project1 := Project{AllocatedCPU: 1, UsedCPU: 1, AllocatedGPU: 1, UsedGPU: 1, AllocatedStorage: 1, UsedStorage: 1}
	project2 := Project{AllocatedCPU: 2, UsedCPU: 2, AllocatedGPU: 2, UsedGPU: 2, AllocatedStorage: 2, UsedStorage: 2}
	projects := make(map[string]Project)
	projects["test_project1"] = project1
	projects["test_project2"] = project2
	executorWithAlloc.Allocations.Projects = projects
	assert.False(t, executor1.Equals(executorWithAlloc))

	executor2 := CreateExecutor(id+"X", executorType, name, colonyName, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType+"X", name, colonyName, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name+"X", colonyName, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyName+"X", commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor2.RequireFuncReg = true
	assert.False(t, executor2.Equals(executor1))
}

func TestIsExecutorArraysEqual(t *testing.T) {
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyName := "test_colony_name"
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor(GenerateRandomID(), executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor2 := CreateExecutor(GenerateRandomID(), executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor3 := CreateExecutor(GenerateRandomID(), executorType, name, colonyName, commissionTime, lastHeardFromTime)
	executor4 := CreateExecutor(GenerateRandomID(), executorType, name, colonyName, commissionTime, lastHeardFromTime)

	var executors1 []*Executor
	executors1 = append(executors1, executor1)
	executors1 = append(executors1, executor2)
	executors1 = append(executors1, executor3)

	var executors2 []*Executor
	executors2 = append(executors2, executor2)
	executors2 = append(executors2, executor3)
	executors2 = append(executors2, executor1)

	var executors3 []*Executor
	executors3 = append(executors3, executor2)
	executors3 = append(executors3, executor3)
	executors3 = append(executors3, executor4)

	var executors4 []*Executor

	assert.True(t, IsExecutorArraysEqual(executors1, executors1))
	assert.True(t, IsExecutorArraysEqual(executors1, executors2))
	assert.False(t, IsExecutorArraysEqual(executors1, executors3))
	assert.False(t, IsExecutorArraysEqual(executors1, executors4))
	assert.True(t, IsExecutorArraysEqual(executors4, executors4))
	assert.True(t, IsExecutorArraysEqual(nil, nil))
	assert.False(t, IsExecutorArraysEqual(nil, executors2))
}

func TestExecutorToJSON(t *testing.T) {
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_executor_type", "test_executor_name", "test_colony_name", commissionTime, lastHeardFromTime)

	jsonString, err := executor1.ToJSON()
	assert.Nil(t, err)

	executor2, err := ConvertJSONToExecutor(jsonString + "error")
	assert.NotNil(t, err)

	executor2, err = ConvertJSONToExecutor(jsonString)
	assert.Nil(t, err)
	assert.True(t, executor2.Equals(executor1))
}

func TestExecutorToJSONArray(t *testing.T) {
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	var executors1 []*Executor
	executors1 = append(executors1, CreateExecutor(GenerateRandomID(), "test_executor_type", "test_executor_name", "test_colony_name", commissionTime, lastHeardFromTime))
	executors1 = append(executors1, CreateExecutor(GenerateRandomID(), "test_executor_type", "test_executor_name", "test_colony_name", commissionTime, lastHeardFromTime))

	jsonString, err := ConvertExecutorArrayToJSON(executors1)
	assert.Nil(t, err)

	executors2, err := ConvertJSONToExecutorArray(jsonString + "error")
	assert.NotNil(t, err)

	executors2, err = ConvertJSONToExecutorArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsExecutorArraysEqual(executors1, executors2))
}
