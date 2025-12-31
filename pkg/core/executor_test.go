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

func TestCreateExecutorFromDB(t *testing.T) {
	id := GenerateRandomID()
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutorFromDB(id, "test_type", "test_name", "test_colony", APPROVED, true, commissionTime, lastHeardFromTime)

	assert.Equal(t, id, executor.ID)
	assert.Equal(t, "test_type", executor.Type)
	assert.Equal(t, "test_name", executor.Name)
	assert.Equal(t, "test_colony", executor.ColonyName)
	assert.Equal(t, APPROVED, executor.State)
	assert.True(t, executor.RequireFuncReg)
}

func TestExecutorApproveRejectUnregister(t *testing.T) {
	executor := CreateExecutor(GenerateRandomID(), "test", "name", "colony", time.Now(), time.Now())

	assert.True(t, executor.IsPending())
	assert.False(t, executor.IsApproved())
	assert.False(t, executor.IsRejected())
	assert.False(t, executor.IsUnregistered())

	executor.Approve()
	assert.True(t, executor.IsApproved())
	assert.False(t, executor.IsPending())

	executor.Reject()
	assert.True(t, executor.IsRejected())
	assert.False(t, executor.IsApproved())

	executor.Unregister()
	assert.True(t, executor.IsUnregistered())
	assert.False(t, executor.IsRejected())
}

func TestIsProjectEqual(t *testing.T) {
	project1 := Project{AllocatedCPU: 10, UsedCPU: 5, AllocatedGPU: 2, UsedGPU: 1, AllocatedStorage: 100, UsedStorage: 50}
	project2 := Project{AllocatedCPU: 10, UsedCPU: 5, AllocatedGPU: 2, UsedGPU: 1, AllocatedStorage: 100, UsedStorage: 50}
	project3 := Project{AllocatedCPU: 20, UsedCPU: 5, AllocatedGPU: 2, UsedGPU: 1, AllocatedStorage: 100, UsedStorage: 50}

	assert.True(t, IsProjectEqual(project1, project2))
	assert.False(t, IsProjectEqual(project1, project3))

	// Test each field difference
	project3 = Project{AllocatedCPU: 10, UsedCPU: 10, AllocatedGPU: 2, UsedGPU: 1, AllocatedStorage: 100, UsedStorage: 50}
	assert.False(t, IsProjectEqual(project1, project3))

	project3 = Project{AllocatedCPU: 10, UsedCPU: 5, AllocatedGPU: 4, UsedGPU: 1, AllocatedStorage: 100, UsedStorage: 50}
	assert.False(t, IsProjectEqual(project1, project3))

	project3 = Project{AllocatedCPU: 10, UsedCPU: 5, AllocatedGPU: 2, UsedGPU: 2, AllocatedStorage: 100, UsedStorage: 50}
	assert.False(t, IsProjectEqual(project1, project3))

	project3 = Project{AllocatedCPU: 10, UsedCPU: 5, AllocatedGPU: 2, UsedGPU: 1, AllocatedStorage: 200, UsedStorage: 50}
	assert.False(t, IsProjectEqual(project1, project3))

	project3 = Project{AllocatedCPU: 10, UsedCPU: 5, AllocatedGPU: 2, UsedGPU: 1, AllocatedStorage: 100, UsedStorage: 100}
	assert.False(t, IsProjectEqual(project1, project3))
}

func TestIsProjectsEqual(t *testing.T) {
	project1 := Project{AllocatedCPU: 10, UsedCPU: 5}
	project2 := Project{AllocatedCPU: 20, UsedCPU: 10}

	projects1 := map[string]Project{"proj1": project1, "proj2": project2}
	projects2 := map[string]Project{"proj1": project1, "proj2": project2}
	projects3 := map[string]Project{"proj1": project1}

	assert.True(t, IsProjectsEqual(projects1, projects2))
	assert.False(t, IsProjectsEqual(projects1, projects3))
}

func TestExecutorEqualsNil(t *testing.T) {
	executor := CreateExecutor(GenerateRandomID(), "test", "name", "colony", time.Now(), time.Now())
	assert.False(t, executor.Equals(nil))
}

func TestExecutorEqualsAllocationsEdgeCases(t *testing.T) {
	executor1 := CreateExecutor(GenerateRandomID(), "test", "name", "colony", time.Now(), time.Now())
	executor2 := CreateExecutor(executor1.ID, "test", "name", "colony", time.Now(), time.Now())

	// One has nil projects, other has non-nil
	executor1.Allocations.Projects = nil
	executor2.Allocations.Projects = map[string]Project{"proj": {AllocatedCPU: 1}}
	assert.False(t, executor1.Equals(executor2))

	// Swap
	executor1.Allocations.Projects = map[string]Project{"proj": {AllocatedCPU: 1}}
	executor2.Allocations.Projects = nil
	assert.False(t, executor1.Equals(executor2))
}

func TestIsHardwareEqualAllFields(t *testing.T) {
	base := Hardware{
		Model: "model", Nodes: 1, CPU: "cpu", Cores: 8, Memory: "16GB",
		Storage: "1TB", Platform: "linux", Architecture: "amd64",
		GPU: GPU{Name: "gpu", Memory: "8GB", Count: 1, NodeCount: 1},
		Network: []string{"192.168.1.1"},
	}

	tests := []struct {
		name   string
		modify func(h *Hardware)
	}{
		{"Model", func(h *Hardware) { h.Model = "different" }},
		{"Nodes", func(h *Hardware) { h.Nodes = 2 }},
		{"CPU", func(h *Hardware) { h.CPU = "different" }},
		{"Memory", func(h *Hardware) { h.Memory = "32GB" }},
		{"Storage", func(h *Hardware) { h.Storage = "2TB" }},
		{"Platform", func(h *Hardware) { h.Platform = "darwin" }},
		{"Architecture", func(h *Hardware) { h.Architecture = "arm64" }},
		{"GPUName", func(h *Hardware) { h.GPU.Name = "different" }},
		{"GPUMemory", func(h *Hardware) { h.GPU.Memory = "16GB" }},
		{"GPUCount", func(h *Hardware) { h.GPU.Count = 2 }},
		{"GPUNodeCount", func(h *Hardware) { h.GPU.NodeCount = 2 }},
		{"NetworkLen", func(h *Hardware) { h.Network = []string{"192.168.1.1", "10.0.0.1"} }},
		{"NetworkVal", func(h *Hardware) { h.Network = []string{"10.0.0.1"} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hw := Hardware{
				Model: "model", Nodes: 1, CPU: "cpu", Cores: 8, Memory: "16GB",
				Storage: "1TB", Platform: "linux", Architecture: "amd64",
				GPU: GPU{Name: "gpu", Memory: "8GB", Count: 1, NodeCount: 1},
				Network: []string{"192.168.1.1"},
			}
			tt.modify(&hw)
			assert.False(t, IsHardwareEqual(base, hw))
		})
	}
}

func TestIsSoftwareEqual(t *testing.T) {
	sw1 := Software{Name: "test", Type: "runtime", Version: "1.0"}
	sw2 := Software{Name: "test", Type: "runtime", Version: "1.0"}
	sw3 := Software{Name: "other", Type: "runtime", Version: "1.0"}
	sw4 := Software{Name: "test", Type: "library", Version: "1.0"}
	sw5 := Software{Name: "test", Type: "runtime", Version: "2.0"}

	assert.True(t, IsSoftwareEqual(sw1, sw2))
	assert.False(t, IsSoftwareEqual(sw1, sw3))
	assert.False(t, IsSoftwareEqual(sw1, sw4))
	assert.False(t, IsSoftwareEqual(sw1, sw5))
}
