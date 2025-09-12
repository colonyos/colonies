package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestProcessClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	process := utils.CreateTestProcess("invalid_id")
	
	// KVStore operations work even after close (in-memory store)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	_, err = db.GetProcesses()
	assert.Nil(t, err)

	// The process we added should be retrievable
	_, err = db.GetProcessByID(process.ID) 
	assert.Nil(t, err)

	_, err = db.FindProcessesByColonyName("invalid_name", 60, core.SUCCESS)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindProcessesByExecutorID("invalid_id", "invalid_id", 60, core.SUCCESS)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindWaitingProcesses("invalid_id", "", "", "", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindRunningProcesses("invalid_id", "", "", "", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindAllRunningProcesses()
	assert.Nil(t, err)

	_, err = db.FindAllWaitingProcesses()
	assert.Nil(t, err)

	_, err = db.FindSuccessfulProcesses("invalid_id", "", "", "", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindFailedProcesses("invalid_id", "", "", "", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindCandidates("invalid_id", "invalid_type", 0, 0, 0, 0, 0, 0, 1)
	assert.Nil(t, err) // Returns empty slice

	err = db.RemoveProcessByID("invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing

	err = db.RemoveAllProcesses()
	assert.Nil(t, err)

	err = db.RemoveAllWaitingProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllRunningProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllSuccessfulProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllFailedProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.RemoveAllProcessesByProcessGraphID("invalid_id")
	assert.Nil(t, err)

	err = db.RemoveAllProcessesInProcessGraphsByColonyName("invalid_name")
	assert.Nil(t, err)

	err = db.ResetProcess(process)
	assert.Nil(t, err)

	input := make([]interface{}, 2)
	input[0] = "result1"
	input[1] = "result2"
	err = db.SetInput("invalid_id", input)
	assert.NotNil(t, err) // Expected error for non-existing process

	output := make([]interface{}, 2)
	output[0] = "result1"
	output[1] = "result2"
	err = db.SetOutput("invalid_id", output)
	assert.NotNil(t, err) // Expected error for non-existing process

	err = db.SetErrors("invalid_id", []string{"error"})
	assert.NotNil(t, err) // Expected error for non-existing process

	err = db.SetProcessState("invalid_id", 1)
	assert.NotNil(t, err) // Expected error for non-existing process

	invalidProcess := &core.Process{ID: "invalid_process"}
	err = db.Assign("invalid_executor", invalidProcess)
	assert.NotNil(t, err) // Expected error for non-existing process

	err = db.Unassign(invalidProcess)
	assert.NotNil(t, err) // Expected error for non-existing process

	err = db.MarkFailed("invalid_process", []string{"error"})
	assert.NotNil(t, err) // Expected error for non-existing process

	_, err = db.CountWaitingProcesses()
	assert.Nil(t, err)

	_, err = db.CountRunningProcesses()
	assert.Nil(t, err)

	_, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)

	_, err = db.CountFailedProcesses()
	assert.Nil(t, err)

	_, err = db.CountWaitingProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountRunningProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountSuccessfulProcessesByColonyName("invalid_name")
	assert.Nil(t, err)

	_, err = db.CountFailedProcessesByColonyName("invalid_name")
	assert.Nil(t, err)
}

func TestAddProcess(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	process := utils.CreateTestProcess("test_colony_name")
	
	// Test adding nil process
	err = db.AddProcess(nil)
	assert.NotNil(t, err)

	// Test adding valid process
	err = db.AddProcess(process)
	assert.Nil(t, err)

	// Test duplicate process
	err = db.AddProcess(process)
	assert.NotNil(t, err)

	// Verify process was added
	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.True(t, process.Equals(processFromDB))

	// Test getting all processes
	processes, err := db.GetProcesses()
	assert.Nil(t, err)
	assert.Len(t, processes, 1)
	assert.True(t, process.Equals(processes[0]))
}

func TestProcessStates(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	process := utils.CreateTestProcess("test_colony_name")
	err = db.AddProcess(process)
	assert.Nil(t, err)

	// Test state transitions
	err = db.SetProcessState(process.ID, core.RUNNING)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.State, core.RUNNING)

	err = db.SetProcessState(process.ID, core.SUCCESS)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.State, core.SUCCESS)

	// Test invalid process ID
	err = db.SetProcessState("invalid_id", core.RUNNING)
	assert.NotNil(t, err)
}

func TestProcessAssignment(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	process := utils.CreateTestProcess("test_colony_name")
	err = db.AddProcess(process)
	assert.Nil(t, err)

	executorID := "test_executor_id"

	// Test assignment
	err = db.Assign(executorID, process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.AssignedExecutorID, executorID)
	assert.Equal(t, processFromDB.State, core.RUNNING)

	// Test unassignment
	err = db.Unassign(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.AssignedExecutorID, "")
	assert.Equal(t, processFromDB.State, core.WAITING)

	// Test invalid process ID
	invalidProcess := &core.Process{ID: "invalid_id"}
	err = db.Assign(executorID, invalidProcess)
	assert.NotNil(t, err)

	err = db.Unassign(invalidProcess)
	assert.NotNil(t, err)
}

func TestProcessInputOutput(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	process := utils.CreateTestProcess("test_colony_name")
	err = db.AddProcess(process)
	assert.Nil(t, err)

	// Test setting input
	input := []interface{}{"input1", "input2", 123}
	err = db.SetInput(process.ID, input)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, len(processFromDB.FunctionSpec.Args), len(input))

	// Test setting output
	output := []interface{}{"output1", "output2", 456}
	err = db.SetOutput(process.ID, output)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, len(processFromDB.Output), len(output))

	// Test setting errors
	errors := []string{"error1", "error2"}
	err = db.SetErrors(process.ID, errors)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.Errors, errors)

	// Test invalid process ID
	err = db.SetInput("invalid_id", input)
	assert.NotNil(t, err)

	err = db.SetOutput("invalid_id", output)
	assert.NotNil(t, err)

	err = db.SetErrors("invalid_id", errors)
	assert.NotNil(t, err)
}

func TestMarkFailed(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	process := utils.CreateTestProcess("test_colony_name")
	process.State = core.RUNNING
	err = db.AddProcess(process)
	assert.Nil(t, err)

	errors := []string{"test error"}
	err = db.MarkFailed(process.ID, errors)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.State, core.FAILED)
	assert.Equal(t, processFromDB.Errors, errors)

	// Test invalid process ID
	err = db.MarkFailed("invalid_id", errors)
	assert.NotNil(t, err)
}

func TestResetProcess(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	process := utils.CreateTestProcess("test_colony_name")
	process.State = core.FAILED
	process.Errors = []string{"some error"}
	err = db.AddProcess(process)
	assert.Nil(t, err)

	// Reset process
	err = db.ResetProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.State, core.WAITING)
	assert.Empty(t, processFromDB.Errors)

	// Test with nil process
	err = db.ResetProcess(nil)
	assert.NotNil(t, err)
}

func TestCountProcesses(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add processes in different states
	process1 := utils.CreateTestProcess("test_colony")
	process1.State = core.WAITING
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess("test_colony")
	process2.State = core.RUNNING
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess("test_colony")
	process3.State = core.SUCCESS
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcess("test_colony")
	process4.State = core.FAILED
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	// Test counts
	waitingCount, err := db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingCount, 1)

	runningCount, err := db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningCount, 1)

	successCount, err := db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successCount, 1)

	failedCount, err := db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedCount, 1)

	// Test colony-specific counts
	waitingByColony, err := db.CountWaitingProcessesByColonyName("test_colony")
	assert.Nil(t, err)
	assert.Equal(t, waitingByColony, 1)

	runningByColony, err := db.CountRunningProcessesByColonyName("test_colony")
	assert.Nil(t, err)
	assert.Equal(t, runningByColony, 1)

	successByColony, err := db.CountSuccessfulProcessesByColonyName("test_colony")
	assert.Nil(t, err)
	assert.Equal(t, successByColony, 1)

	failedByColony, err := db.CountFailedProcessesByColonyName("test_colony")
	assert.Nil(t, err)
	assert.Equal(t, failedByColony, 1)
}

func TestRemoveProcesses(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add test processes
	process1 := utils.CreateTestProcess("test_colony")
	process1.State = core.WAITING
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess("test_colony")
	process2.State = core.RUNNING
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	// Test remove by ID
	err = db.RemoveProcessByID(process1.ID)
	assert.Nil(t, err)

	_, err = db.GetProcessByID(process1.ID)
	assert.NotNil(t, err)

	// Test remove invalid ID
	err = db.RemoveProcessByID("invalid_id")
	assert.NotNil(t, err)

	// Verify other process still exists
	_, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)

	// Test remove all
	err = db.RemoveAllProcesses()
	assert.Nil(t, err)

	processes, err := db.GetProcesses()
	assert.Nil(t, err)
	assert.Empty(t, processes)
}