package postgresql

import (
	"colonies/pkg/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddProcess(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	runtime1ID := core.GenerateRandomID()
	runtime2ID := core.GenerateRandomID()

	processSpec := core.CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process := core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.True(t, process.Equals(processFromDB))
	assert.Contains(t, processFromDB.ProcessSpec.Conditions.RuntimeIDs, runtime1ID)
	assert.Contains(t, processFromDB.ProcessSpec.Conditions.RuntimeIDs, runtime2ID)
}

func TestAddProcessWithEnv(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	env := make(map[string]string)
	env["test_key_1"] = "test_value_1"
	env["test_key_2"] = "test_value_2"

	colonyID := core.GenerateRandomID()
	processSpec := core.CreateProcessSpec(colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, env)
	process := core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.True(t, process.Equals(processFromDB))
}

func TestDeleteProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	runtime1ID := core.GenerateRandomID()
	runtime2ID := core.GenerateRandomID()

	processSpec1 := core.CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	processSpec2 := core.CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processSpec3 := core.CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process3 := core.CreateProcess(processSpec3)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	numberOfProcesses, err := db.NumberOfWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	err = db.DeleteProcessByID(process1.ID)
	assert.Nil(t, err)

	numberOfProcesses, err = db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	err = db.DeleteAllProcesses()
	assert.Nil(t, err)

	numberOfProcesses, err = db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfProcesses)
}

func TestDeleteAllProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()

	processSpec1 := core.CreateProcessSpec(colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteAllProcesses()
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)
}

func TestDeleteProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()

	processSpec1 := core.CreateProcessSpec(colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	processSpec2 := core.CreateProcessSpec(colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(process2.ID, core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteProcessByID(process1.ID)
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttribute(process2.ID, "test_key2", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Not deleted as it belongs to process 2
}

func TestAssign(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	runtime := core.CreateRuntime(colony.ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process := core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, processFromDB.State)
	assert.False(t, processFromDB.IsAssigned)

	err = db.AssignRuntime(runtime.ID, process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.True(t, processFromDB.IsAssigned)
	assert.False(t, int64(processFromDB.StartTime.Sub(processFromDB.SubmissionTime)) < 0)
	assert.Equal(t, core.RUNNING, processFromDB.State)

	err = db.UnassignRuntime(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.False(t, processFromDB.IsAssigned)
	assert.False(t, int64(processFromDB.EndTime.Sub(processFromDB.StartTime)) < 0)
}

func TestMarkSuccessful(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	runtime := core.CreateRuntime(colony.ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process := core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, process.State)

	err = db.MarkSuccessful(process)
	assert.NotNil(t, err) // Not possible to set waiting process to successfull

	err = db.AssignRuntime(runtime.ID, process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, process.State)

	err = db.MarkSuccessful(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.SUCCESS, processFromDB.State)

	err = db.MarkFailed(process)
	assert.NotNil(t, err) // Not possible to set successful process as failed
}

func TestMarkFailed(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	runtime := core.CreateRuntime(colony.ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process := core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, process.State)

	err = db.MarkFailed(process)
	assert.NotNil(t, err) // Not possible to set waiting process to failed

	err = db.AssignRuntime(runtime.ID, process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, processFromDB.State)

	err = db.MarkFailed(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.FAILED, processFromDB.State)

	err = db.MarkFailed(process)
	assert.NotNil(t, err) // Not possible to set successful process as failed
}

func TestReset(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	runtime := core.CreateRuntime(colony.ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process := core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process)
	assert.Nil(t, err)

	processSpec = core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process = core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process)
	assert.Nil(t, err)

	processSpec = core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process = core.CreateProcess(processSpec)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.AssignRuntime(runtime.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err := db.NumberOfFailedProcesses()
	assert.Equal(t, 3, numberOfFailedProcesses)

	err = db.ResetProcess(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err = db.NumberOfFailedProcesses()
	assert.Equal(t, 2, numberOfFailedProcesses)

	err = db.ResetAllProcesses(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err = db.NumberOfFailedProcesses()
	assert.Equal(t, 0, numberOfFailedProcesses)
}

func TestFindUnassignedProcesses1(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1ID := core.GenerateRandomID()
	runtime1 := core.CreateRuntime(runtime1ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2ID := core.GenerateRandomID()
	runtime2 := core.CreateRuntime(runtime2ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	processSpec1 := core.CreateProcessSpec(colony.ID, []string{runtime2.ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(colony.ID, []string{runtime2.ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, runtime2.ID, runtime2.RuntimeType, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID, runtime2.ID, runtime2.RuntimeType, 2)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 2)
}

func TestFindUnassignedProcesses2(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1ID := core.GenerateRandomID()
	runtime1 := core.CreateRuntime(runtime1ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2ID := core.GenerateRandomID()
	runtime2 := core.CreateRuntime(runtime2ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	processSpec1 := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(colony.ID, []string{runtime2.ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processSpec3 := core.CreateProcessSpec(colony.ID, []string{runtime2.ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process3 := core.CreateProcess(processSpec3)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, runtime2.ID, runtime2.RuntimeType, 2)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 2)

	counter := 0
	for _, processFromDB := range processsFromDB {
		if processFromDB.ID == process1.ID {
			counter++
		}

		if processFromDB.ID == process2.ID {
			counter++
		}
	}

	assert.Equal(t, 2, counter)
}

// Test that the order of targetRuntimeIDs strings does not matter
func TestFindUnassignedProcesses3(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1ID := core.GenerateRandomID()
	runtime1 := core.CreateRuntime(runtime1ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2ID := core.GenerateRandomID()
	runtime2 := core.CreateRuntime(runtime2ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	processSpec1 := core.CreateProcessSpec(colony.ID, []string{runtime1.ID, runtime2.ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(colony.ID, []string{runtime1.ID, runtime2.ID}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, runtime1.ID, runtime1.RuntimeType, 1)
	assert.Nil(t, err)

	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID, runtime2.ID, runtime2.RuntimeType, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)
}

// Test that runtime type matching is working
func TestFindUnassignedProcesses4(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1ID := core.GenerateRandomID()
	runtime1 := core.CreateRuntime(runtime1ID, "test_runtime_type_1", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2ID := core.GenerateRandomID()
	runtime2 := core.CreateRuntime(runtime2ID, "test_runtime_type_2", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	processSpec1 := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type_1", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type_2", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, runtime1.ID, runtime1.RuntimeType, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID, runtime2.ID, runtime2.RuntimeType, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process2.ID, processsFromDB[0].ID)
}

func TestFindProcessAssigned(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtime1ID := core.GenerateRandomID()
	runtime1 := core.CreateRuntime(runtime1ID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	processSpec1 := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process1 := core.CreateProcess(processSpec1)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	process2 := core.CreateProcess(processSpec2)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	numberOfProcesses, err := db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	numberOfRunningProcesses, err := db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfRunningProcesses)

	numberOfSuccesfulProcesses, err := db.NumberOfSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfSuccesfulProcesses)

	numberOfFailedProcesses, err := db.NumberOfFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfFailedProcesses)

	processsFromDB1, err := db.FindUnassignedProcesses(colony.ID, runtime1.ID, runtime1.RuntimeType, 1)
	assert.Nil(t, err)
	assert.Equal(t, process1.ID, processsFromDB1[0].ID)
	assert.Len(t, processsFromDB1, 1)

	err = db.AssignRuntime(runtime1.ID, processsFromDB1[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfRunningProcesses)

	processsFromDB2, err := db.FindUnassignedProcesses(colony.ID, runtime1.ID, runtime1.RuntimeType, 1)
	assert.Nil(t, err)
	assert.Equal(t, process2.ID, processsFromDB2[0].ID)

	err = db.AssignRuntime(runtime1.ID, processsFromDB2[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfRunningProcesses)

	err = db.MarkSuccessful(processsFromDB1[0])
	assert.Nil(t, err)

	err = db.MarkFailed(processsFromDB2[0])
	assert.Nil(t, err)

	numberOfSuccesfulProcesses, err = db.NumberOfSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfSuccesfulProcesses)

	numberOfFailedProcesses, err = db.NumberOfFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfFailedProcesses)
}

func TestFindWaitingProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	runtimeID := core.GenerateRandomID()
	runtime := core.CreateRuntime(runtimeID, "test_runtime_type", "test_runtime_name", colony.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddRuntime(runtime)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	waitingProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		process := core.CreateProcess(processSpec)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		//	err = db.AssignRuntime(runtimeID, process)
		//assert.Nil(t, err)
		waitingProcessIDs[process.ID] = true
	}
	waitingProcessIDsFromDB, err := db.FindWaitingProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Create some running processes
	runningProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		process := core.CreateProcess(processSpec)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.AssignRuntime(runtimeID, process)
		assert.Nil(t, err)
		runningProcessIDs[process.ID] = true
	}
	runningProcessIDsFromDB, err := db.FindRunningProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Create some successful processes
	successfulProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		process := core.CreateProcess(processSpec)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.AssignRuntime(runtimeID, process)
		assert.Nil(t, err)
		err = db.MarkSuccessful(process)
		assert.Nil(t, err)
		successfulProcessIDs[process.ID] = true
	}
	successfulProcessIDsFromDB, err := db.FindSuccessfulProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Create some successful processes
	failedProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		processSpec := core.CreateProcessSpec(colony.ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		process := core.CreateProcess(processSpec)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.AssignRuntime(runtimeID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process)
		assert.Nil(t, err)
		failedProcessIDs[process.ID] = true
	}
	failedProcessIDsFromDB, err := db.FindFailedProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Now, lets to some checks
	counter := 0
	for _, processFromDB := range waitingProcessIDsFromDB {
		if waitingProcessIDs[processFromDB.ID] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	counter = 0
	for _, processFromDB := range runningProcessIDsFromDB {
		if runningProcessIDs[processFromDB.ID] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	counter = 0
	for _, processFromDB := range successfulProcessIDsFromDB {
		if successfulProcessIDs[processFromDB.ID] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	counter = 0
	for _, processFromDB := range failedProcessIDsFromDB {
		if failedProcessIDs[processFromDB.ID] {
			counter++
		}
	}
	assert.Equal(t, 10, counter)

	numberOfProcesses, err := db.NumberOfProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 40, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.NumberOfFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)
}
