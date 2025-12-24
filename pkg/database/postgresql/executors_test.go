package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestExecutorClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	executor := utils.CreateTestExecutor(core.GenerateRandomID())
	err = db.AddExecutor(executor)
	assert.NotNil(t, err)

	_, err = db.GetExecutors()
	assert.NotNil(t, err)

	_, err = db.GetExecutorByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetExecutorsByColonyName("invalid_colony_name")
	assert.NotNil(t, err)

	_, err = db.GetExecutorByName("invalid_id", "invalid_name")
	assert.NotNil(t, err)

	err = db.ApproveExecutor(executor)
	assert.NotNil(t, err)

	err = db.RejectExecutor(executor)
	assert.NotNil(t, err)

	err = db.MarkAlive(executor)
	assert.NotNil(t, err)

	err = db.RemoveExecutorByName("invalid_colony_name", "invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveExecutorsByColonyName("invalid_colony_name")
	assert.NotNil(t, err)

	_, err = db.CountExecutors()
	assert.NotNil(t, err)

	_, err = db.CountExecutorsByColonyName("invalid_colony_name")
	assert.NotNil(t, err)
}

func TestAddExecutor(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	executor.Capabilities.Software[0].Name = "sw_name"
	executor.Capabilities.Software[0].Type = "sw_type"
	executor.Capabilities.Software[0].Version = "sw_version"

	executor.Capabilities.Hardware[0].Model = "model"
	executor.Capabilities.Hardware[0].Nodes = 10
	executor.Capabilities.Hardware[0].CPU = "1000m"
	executor.Capabilities.Hardware[0].Memory = "10G"
	executor.Capabilities.Hardware[0].Storage = "1000G"
	executor.Capabilities.Hardware[0].GPU.Name = "nvidia_2080ti"
	executor.Capabilities.Hardware[0].GPU.Count = 4000
	executor.Capabilities.Hardware[0].GPU.NodeCount = 4
	executor.Capabilities.Hardware[0].GPU.Memory = "10G"

	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executors, err := db.GetExecutors()
	assert.Nil(t, err)
	assert.Len(t, executors, 1)

	executorFromDB := executors[0]
	assert.True(t, executor.Equals(executorFromDB))
	assert.True(t, executorFromDB.IsPending())
	assert.False(t, executorFromDB.IsApproved())
	assert.False(t, executorFromDB.IsRejected())

	assert.Len(t, executor.Capabilities.Software, 1)
	assert.Equal(t, executor.Capabilities.Software[0].Name, "sw_name")
	assert.Equal(t, executor.Capabilities.Software[0].Type, "sw_type")
	assert.Equal(t, executor.Capabilities.Software[0].Version, "sw_version")

	assert.Len(t, executor.Capabilities.Hardware, 1)
	assert.Equal(t, executor.Capabilities.Hardware[0].Model, "model")
	assert.Equal(t, executor.Capabilities.Hardware[0].Nodes, 10)
	assert.Equal(t, executor.Capabilities.Hardware[0].CPU, "1000m")
	assert.Equal(t, executor.Capabilities.Hardware[0].Memory, "10G")
	assert.Equal(t, executor.Capabilities.Hardware[0].Storage, "1000G")
	assert.Equal(t, executor.Capabilities.Hardware[0].GPU.Name, "nvidia_2080ti")
	assert.Equal(t, executor.Capabilities.Hardware[0].GPU.Count, 4000)
	assert.Equal(t, executor.Capabilities.Hardware[0].GPU.NodeCount, 4)
	assert.Equal(t, executor.Capabilities.Hardware[0].GPU.Memory, "10G")
}

func TestAddExecutorWithLocation(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	executor.LocationName = "Home"

	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromDB)
	assert.Equal(t, "Home", executorFromDB.LocationName)
}

func TestAddDuplicateExecutorRejected(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Create and add first executor
	executor1 := utils.CreateTestExecutor(colony.Name)
	executor1.Name = "test-executor-same-name"
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	// Try to add second executor with same name - should be rejected
	executor2 := utils.CreateTestExecutor(colony.Name)
	executor2.Name = "test-executor-same-name"
	executor2.ID = core.GenerateRandomID() // Different ID, same name
	err = db.AddExecutor(executor2)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify only one executor exists
	executors, err := db.GetExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, executors, 1)
	assert.Equal(t, executor1.ID, executors[0].ID)
}

func TestAddDuplicateExecutorConcurrentRejected(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Launch 10 concurrent attempts to add executor with same name
	const numGoroutines = 10
	results := make(chan error, numGoroutines)
	executorName := "concurrent-test-executor"

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			executor := utils.CreateTestExecutor(colony.Name)
			executor.Name = executorName
			executor.ID = core.GenerateRandomID()
			results <- db.AddExecutor(executor)
		}(i)
	}

	// Collect results
	successCount := 0
	failureCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			failureCount++
			// All failures should be "already exists" errors
			assert.Contains(t, err.Error(), "already exists")
		}
	}

	// Exactly one should succeed, rest should fail
	assert.Equal(t, 1, successCount, "Exactly one executor should be added successfully")
	assert.Equal(t, numGoroutines-1, failureCount, "All other attempts should fail with duplicate error")

	// Verify only one executor exists in database
	executors, err := db.GetExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Len(t, executors, 1)
	assert.Equal(t, executorName, executors[0].Name)
}

func TestSameExecutorNameDifferentColoniesAllowed(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	// Create two colonies
	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	// Add executor with same name to both colonies - should succeed
	executor1 := utils.CreateTestExecutor(colony1.Name)
	executor1.Name = "shared-executor-name"
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony2.Name)
	executor2.Name = "shared-executor-name"
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	// Verify both executors exist
	executorsColony1, err := db.GetExecutorsByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.Len(t, executorsColony1, 1)

	executorsColony2, err := db.GetExecutorsByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.Len(t, executorsColony2, 1)
}

func TestAddExecutorWithAllocations(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	project := core.Project{AllocatedCPU: 1, UsedCPU: 2, AllocatedGPU: 3, UsedGPU: 4, AllocatedStorage: 5, UsedStorage: 6}
	projects := make(map[string]core.Project)
	projects["test_project"] = project
	executor.Allocations.Projects = projects

	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executors, err := db.GetExecutors()
	assert.Nil(t, err)
	assert.Len(t, executors, 1)
	testProj := executors[0].Allocations.Projects["test_project"]
	assert.Equal(t, testProj.AllocatedCPU, int64(1))
	assert.Equal(t, testProj.UsedCPU, int64(2))
	assert.Equal(t, testProj.AllocatedGPU, int64(3))
	assert.Equal(t, testProj.UsedGPU, int64(4))
	assert.Equal(t, testProj.AllocatedStorage, int64(5))
	assert.Equal(t, testProj.UsedStorage, int64(6))
}

func TestSetAllocations(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	project := core.Project{AllocatedCPU: 1, UsedCPU: 2, AllocatedGPU: 3, UsedGPU: 4, AllocatedStorage: 5, UsedStorage: 6}
	projects := make(map[string]core.Project)
	projects["test_project"] = project
	executor.Allocations.Projects = projects

	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executors, err := db.GetExecutors()
	assert.Nil(t, err)
	assert.Len(t, executors, 1)
	testProj := executors[0].Allocations.Projects["test_project"]
	assert.Equal(t, testProj.AllocatedCPU, int64(1))
	assert.Equal(t, testProj.UsedCPU, int64(2))
	assert.Equal(t, testProj.AllocatedGPU, int64(3))
	assert.Equal(t, testProj.UsedGPU, int64(4))
	assert.Equal(t, testProj.AllocatedStorage, int64(5))
	assert.Equal(t, testProj.UsedStorage, int64(6))

	project = core.Project{AllocatedCPU: 7, UsedCPU: 8, AllocatedGPU: 9, UsedGPU: 10, AllocatedStorage: 11, UsedStorage: 12}
	projects = make(map[string]core.Project)
	projects["test_project"] = project
	allocations := core.Allocations{Projects: projects}

	err = db.SetAllocations(colony.Name, executor.Name, allocations)
	assert.Nil(t, err)

	executors, err = db.GetExecutors()
	assert.Nil(t, err)
	assert.Len(t, executors, 1)
	testProj = executors[0].Allocations.Projects["test_project"]
	assert.Equal(t, testProj.AllocatedCPU, int64(7))
	assert.Equal(t, testProj.UsedCPU, int64(8))
	assert.Equal(t, testProj.AllocatedGPU, int64(9))
	assert.Equal(t, testProj.UsedGPU, int64(10))
	assert.Equal(t, testProj.AllocatedStorage, int64(11))
	assert.Equal(t, testProj.UsedStorage, int64(12))
}

func TestAddExecutors(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	err = db.AddExecutor(nil)
	assert.NotNil(t, err) // Error

	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	err = db.AddExecutor(executor1) // Try to add the same executor again
	assert.NotNil(t, err)           // Error

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony.Name)
	executor3.Name = executor2.Name // Note name not unique
	err = db.AddExecutor(executor3)
	assert.NotNil(t, err) // Error

	executor3 = utils.CreateTestExecutor(colony.Name)
	executor3.Name = "unique_name"
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	var executors []*core.Executor
	executors = append(executors, executor1)
	executors = append(executors, executor2)
	executors = append(executors, executor3)

	executorsFromDB, err := db.GetExecutors()
	assert.Nil(t, err)
	assert.True(t, core.IsExecutorArraysEqual(executors, executorsFromDB))
}

func TestGetExecutorByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID("invalid_id")
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByID(executor1.ID)
	assert.Nil(t, err)
	assert.True(t, executor1.Equals(executorFromDB))
}

func TestGetExecutorByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)
	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	assert.Nil(t, err)

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.Name)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	var executorsColony1 []*core.Executor
	executorsColony1 = append(executorsColony1, executor1)
	executorsColony1 = append(executorsColony1, executor2)

	executorsColony1FromDB, err := db.GetExecutorsByColonyName("invalid_colony_name")
	assert.Nil(t, err)
	assert.NotNil(t, executorsColony1)

	executorsColony1FromDB, err = db.GetExecutorsByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.True(t, core.IsExecutorArraysEqual(executorsColony1, executorsColony1FromDB))
}

func TestGetExecutorByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	executor1.Name = "test_name_1"
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	executor2.Name = "test_name_"
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByName("invalid__id", executor1.Name)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByName(colony.Name, "invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByName("invalid__id", "invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByName(colony.Name, executor1.Name)
	assert.Nil(t, err)
	assert.True(t, executor1.Equals(executorFromDB))
}

func TestMarkAlive(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	time.Sleep(3000 * time.Millisecond)

	err = db.MarkAlive(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)

	assert.True(t, (executorFromDB.LastHeardFromTime.Unix()-executor.LastHeardFromTime.Unix()) > 1)
}

func TestApproveExecutor(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	assert.True(t, executor.IsPending())

	err = db.ApproveExecutor(executor)
	assert.Nil(t, err)

	assert.False(t, executor.IsPending())
	assert.False(t, executor.IsRejected())
	assert.True(t, executor.IsApproved())

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.True(t, executorFromDB.IsApproved())

	err = db.RejectExecutor(executor)
	assert.Nil(t, err)
	assert.True(t, executor.IsRejected())

	executorFromDB, err = db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.True(t, executor.IsRejected())
}

func TestRemoveExecutorMoveBackToQueue(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor1.Name, ColonyName: colony.Name, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor2.Name, ColonyName: colony.Name, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	env := make(map[string]string)

	process1 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process4.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	err = db.Assign(executor1.ID, process1)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process2)
	assert.Nil(t, err)
	err = db.Assign(executor2.ID, process3)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process4)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor1.ID)

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor1.ID)

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor2.ID)

	count, err := db.CountWaitingProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 0)

	_, _, err = db.MarkSuccessful(process4.ID)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colony.Name)
	assert.Len(t, functions, 2)

	err = db.RemoveExecutorByName(colony.Name, executor1.Name)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colony.Name)
	assert.Len(t, functions, 1)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor2.ID)

	count, err = db.CountWaitingProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 2)

	count, err = db.CountSuccessfulProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 1)

	count, err = db.CountRunningProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 1)

	count, err = db.CountFailedProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 0)
}

func TestRemoveExecutorsMoveBackToQueue(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	env := make(map[string]string)

	process1 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithEnv(colony.Name, env)
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process4.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	err = db.Assign(executor1.ID, process1)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process2)
	assert.Nil(t, err)
	err = db.Assign(executor2.ID, process3)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process4)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor1.ID)

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor1.ID)

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == executor2.ID)

	count, err := db.CountWaitingProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 0)

	_, _, err = db.MarkSuccessful(process4.ID)
	assert.Nil(t, err)

	err = db.RemoveExecutorsByColonyName(colony.Name)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	processFromDB, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.True(t, processFromDB.AssignedExecutorID == "")

	count, err = db.CountWaitingProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 3)

	count, err = db.CountSuccessfulProcessesByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, count == 1)
}

func TestRemoveExecutors(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor1.Name, ColonyName: colony1.Name, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor2.Name, ColonyName: colony1.Name, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.Name)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executor3.Name, ColonyName: colony2.Name, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colony1.Name)
	assert.Len(t, functions, 2)

	functions, err = db.GetFunctionsByColonyName(colony2.Name)
	assert.Len(t, functions, 1)

	err = db.RemoveExecutorByName(colony1.Name, executor2.Name)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor2.ID)
	assert.Nil(t, err)

	// After RemoveExecutorByName, executor should still exist but be UNREGISTERED (soft delete)
	assert.NotNil(t, executorFromDB)
	assert.Equal(t, core.UNREGISTERED, executorFromDB.State)

	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executorFromDB, err = db.GetExecutorByID(executor2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromDB)

	err = db.RemoveExecutorsByColonyName(colony1.Name)
	assert.Nil(t, err)

	executorFromDB, err = db.GetExecutorByID(executor1.ID)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByID(executor2.ID)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByID(executor3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromDB)

	functions, err = db.GetFunctionsByColonyName(colony1.Name)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyName(colony2.Name)
	assert.Len(t, functions, 1)
}

func TestCountExecutors(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	executorCount, err := db.CountExecutors()
	assert.Nil(t, err)
	assert.True(t, executorCount == 0)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorCount, err = db.CountExecutors()
	assert.Nil(t, err)
	assert.True(t, executorCount == 1)
}

func TestCountExectorsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executor = utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor = utils.CreateTestExecutor(colony2.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorCount, err := db.CountExecutors()
	assert.Nil(t, err)
	assert.True(t, executorCount == 3)

	executorCount, err = db.CountExecutorsByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.True(t, executorCount == 2)

	executorCount, err = db.CountExecutorsByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.True(t, executorCount == 1)

}

func TestChangeExecutorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyName := core.GenerateRandomID()

	executor := utils.CreateTestExecutor(colonyName)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByName(colonyName, executor.Name)
	assert.Nil(t, err)

	err = db.ChangeExecutorID(colonyName, executor.ID, "new_id")
	assert.Nil(t, err)

	executorFromDB, err = db.GetExecutorByName(colonyName, executor.Name)
	assert.Nil(t, err)
	assert.Equal(t, "new_id", executorFromDB.ID)
	assert.NotEqual(t, executorFromDB.ID, executor.ID)

	defer db.Close()
}

