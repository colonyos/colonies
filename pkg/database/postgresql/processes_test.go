package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestProcessClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	process := utils.CreateTestProcess("invalid_id")
	err = db.AddProcess(process)
	assert.NotNil(t, err)

	_, err = db.GetProcesses()
	assert.NotNil(t, err)

	_, err = db.GetProcessByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.FindProcessesByColonyID("invalid_id", 60, core.SUCCESS)
	assert.NotNil(t, err)

	_, err = db.FindProcessesByExecutorID("invalid_id", "invalid_id", 60, core.SUCCESS)
	assert.NotNil(t, err)

	_, err = db.FindWaitingProcesses("invalid_id", 1)
	assert.NotNil(t, err)

	_, err = db.FindRunningProcesses("invalid_id", 1)
	assert.NotNil(t, err)

	_, err = db.FindAllRunningProcesses()
	assert.NotNil(t, err)

	_, err = db.FindAllWaitingProcesses()
	assert.NotNil(t, err)

	_, err = db.FindSuccessfulProcesses("invalid_id", 1)
	assert.NotNil(t, err)

	_, err = db.FindFailedProcesses("invalid_id", 1)
	assert.NotNil(t, err)

	_, err = db.FindUnassignedProcesses("invalid_id", "invalid_id", "invalid_type", 1)
	assert.NotNil(t, err)

	err = db.DeleteProcessByID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllProcesses()
	assert.NotNil(t, err)

	err = db.DeleteAllWaitingProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllRunningProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllSuccessfulProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllFailedProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllProcessesByProcessGraphID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteAllProcessesInProcessGraphsByColonyID("invalid_id")
	assert.NotNil(t, err)

	err = db.ResetProcess(process)
	assert.NotNil(t, err)

	input := make([]interface{}, 2)
	input[0] = "result1"
	input[1] = "result2"
	err = db.SetInput("invalid_id", input)
	assert.NotNil(t, err)

	output := make([]interface{}, 2)
	output[0] = "result1"
	output[1] = "result2"
	err = db.SetOutput("invalid_id", output)
	assert.NotNil(t, err)

	err = db.SetErrors("invalid_id", []string{"error"})
	assert.NotNil(t, err)

	err = db.SetProcessState("invalid_id", 1)
	assert.NotNil(t, err)

	parent := core.GenerateRandomID()
	parents := []string{parent}
	err = db.SetParents("invalid_id", parents)
	assert.NotNil(t, err)

	child := core.GenerateRandomID()
	children := []string{child}
	err = db.SetChildren("invalid_id", children)
	assert.NotNil(t, err)

	err = db.SetWaitForParents("invalid_id", false)
	assert.NotNil(t, err)

	err = db.Assign("invalid_id", process)
	assert.NotNil(t, err)

	err = db.Unassign(process)
	assert.NotNil(t, err)

	_, _, err = db.MarkSuccessful("invalid_id")
	assert.NotNil(t, err)

	err = db.MarkFailed("invalid_id", []string{"error"})
	assert.NotNil(t, err)

	_, err = db.CountProcesses()
	assert.NotNil(t, err)
	_, err = db.CountWaitingProcesses()
	assert.NotNil(t, err)
	_, err = db.CountRunningProcesses()
	assert.NotNil(t, err)
	_, err = db.CountSuccessfulProcesses()
	assert.NotNil(t, err)
	_, err = db.CountFailedProcesses()
	assert.NotNil(t, err)
	_, err = db.CountWaitingProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)
	_, err = db.CountRunningProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)
	_, err = db.CountSuccessfulProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)
	_, err = db.CountFailedProcessesByColonyID("invalid_id")
	assert.NotNil(t, err)
}

func TestAddProcess(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process := utils.CreateTestProcessWithTargets(colonyID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Contains(t, processFromDB.FunctionSpec.Conditions.ExecutorIDs, executor1ID)
	assert.Contains(t, processFromDB.FunctionSpec.Conditions.ExecutorIDs, executor2ID)
}

func TestSelectCandiate(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	var processes []*core.Process

	selectedProcess := db.selectCandidate(processes)
	assert.Nil(t, selectedProcess)

	process1 := utils.CreateTestProcess(core.GenerateRandomID())
	process2 := utils.CreateTestProcess(core.GenerateRandomID())
	processes = append(processes, process1)
	processes = append(processes, process2)

	selectedProcess = db.selectCandidate(processes)
	assert.NotNil(t, selectedProcess)
	assert.Equal(t, selectedProcess.ID, process1.ID)
}

func TestAddProcessWithEnv(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	env := make(map[string]string)
	env["test_key_1"] = "test_value_1"
	env["test_key_2"] = "test_value_2"

	colonyID := core.GenerateRandomID()
	process := utils.CreateTestProcessWithEnv(colonyID, env)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	_, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	processesFromDB, err := db.GetProcesses()
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 1)
}

func TestDeleteProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process1 := utils.CreateTestProcessWithTargets(colonyID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithTargets(colonyID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colonyID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	numberOfProcesses, err := db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	numberOfProcesses, err = db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	err = db.DeleteProcessByID(process1.ID)
	assert.Nil(t, err)

	numberOfProcesses, err = db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	err = db.DeleteAllProcesses()
	assert.Nil(t, err)

	numberOfProcesses, err = db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfProcesses)
}

func TestDeleteAllProcessesByColony(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1ID := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colony1ID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colony1ID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	colony2ID := core.GenerateRandomID()
	process2 := utils.CreateTestProcess(colony2ID)
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colony2ID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	err = db.DeleteAllProcessesByColonyID(colony2ID)
	assert.Nil(t, err)

	_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.Nil(t, err)
	_, err = db.GetAttribute(process2.ID, "test_key1", core.IN)
	assert.NotNil(t, err)
}

func TestDeleteAllProcessesByColonyWithState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1ID := core.GenerateRandomID()
	colony2ID := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process1 := utils.CreateTestProcessWithTargets(colony1ID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithTargets(colony1ID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colony1ID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithTargets(colony1ID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	process5 := utils.CreateTestProcessWithTargets(colony2ID, []string{executor1ID, executor2ID})
	err = db.AddProcess(process5)
	assert.Nil(t, err)

	err = db.SetProcessState(process1.ID, core.WAITING)
	assert.Nil(t, err)

	err = db.SetProcessState(process2.ID, core.RUNNING)
	assert.Nil(t, err)

	err = db.SetProcessState(process3.ID, core.SUCCESS)
	assert.Nil(t, err)

	err = db.SetProcessState(process4.ID, core.FAILED)
	assert.Nil(t, err)

	err = db.SetProcessState(process5.ID, core.FAILED)
	assert.Nil(t, err)

	waitingProcesses, err := db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingProcesses, 1)
	runningProcesses, err := db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 1)
	successfulProcesses, err := db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successfulProcesses, 1)
	failedProcesses, err := db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedProcesses, 2)

	err = db.DeleteAllWaitingProcessesByColonyID(colony1ID)
	waitingProcesses, err = db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingProcesses, 0)
	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 1)
	successfulProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successfulProcesses, 1)
	failedProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedProcesses, 2)

	err = db.DeleteAllRunningProcessesByColonyID(colony1ID)
	waitingProcesses, err = db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingProcesses, 0)
	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 0)
	successfulProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successfulProcesses, 1)
	failedProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedProcesses, 2)

	err = db.DeleteAllSuccessfulProcessesByColonyID(colony1ID)
	waitingProcesses, err = db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingProcesses, 0)
	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 0)
	successfulProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successfulProcesses, 0)
	failedProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedProcesses, 2)

	err = db.DeleteAllFailedProcessesByColonyID(colony1ID)
	waitingProcesses, err = db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingProcesses, 0)
	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 0)
	successfulProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successfulProcesses, 0)
	failedProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedProcesses, 1)

	err = db.DeleteAllFailedProcessesByColonyID(colony2ID)
	waitingProcesses, err = db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, waitingProcesses, 0)
	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 0)
	successfulProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, successfulProcesses, 0)
	failedProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, failedProcesses, 0)
}

func TestDeleteAllProcessesByProcessGraphID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	processGraphID := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyID)
	process1.ProcessGraphID = processGraphID
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colonyID, processGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyID)
	process2.ProcessGraphID = processGraphID
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colonyID, processGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colonyID)
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	attribute3 := core.CreateAttribute(process3.ID, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	processFromServer, err := db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)

	err = db.DeleteAllProcessesByProcessGraphID(processGraphID)
	assert.Nil(t, err)

	processFromServer, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.Nil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.Nil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)
}

func TestDeleteAllProcessesInProcessGraphsByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyID)
	process1.ProcessGraphID = processGraphID1
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colonyID, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyID)
	process2.ProcessGraphID = processGraphID2
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colonyID, processGraphID2, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colonyID)
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	attribute3 := core.CreateAttribute(process3.ID, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	processFromServer, err := db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)

	err = db.DeleteAllProcessesInProcessGraphsByColonyID(colonyID)
	assert.Nil(t, err)

	processFromServer, err = db.GetProcessByID(process1.ID)
	assert.Nil(t, err)
	assert.Nil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process2.ID)
	assert.Nil(t, err)
	assert.Nil(t, processFromServer)

	processFromServer, err = db.GetProcessByID(process3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, processFromServer)
}

func TestDeleteAllProcessesInProcessGraphsByColonyIDWithState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()
	processGraphID3 := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyID)
	process1.ProcessGraphID = processGraphID1
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colonyID, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyID)
	process2.ProcessGraphID = processGraphID2
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colonyID, processGraphID2, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colonyID)
	process3.ProcessGraphID = processGraphID3
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	attribute3 := core.CreateAttribute(process3.ID, colonyID, processGraphID3, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcess(colonyID)
	err = db.AddProcess(process4)
	assert.Nil(t, err)
	attribute4 := core.CreateAttribute(process4.ID, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute4)
	assert.Nil(t, err)

	err = db.SetProcessState(process1.ID, core.FAILED)
	assert.Nil(t, err)

	err = db.SetProcessState(process2.ID, core.FAILED)
	assert.Nil(t, err)

	err = db.SetProcessState(process3.ID, core.RUNNING)
	assert.Nil(t, err)

	err = db.SetProcessState(process4.ID, core.RUNNING)
	assert.Nil(t, err)

	runningProcesses, err := db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 2)

	err = db.DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID, core.FAILED)
	assert.Nil(t, err)

	runningProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 0)

	err = db.DeleteAllProcessesInProcessGraphsByColonyIDWithState(colonyID, core.RUNNING)
	assert.Nil(t, err)

	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 1)
}

func TestDeleteAllProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID, colonyID, core.GenerateRandomID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteAllProcesses()
	assert.Nil(t, err)

	_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.NotNil(t, err)
}

func TestDeleteProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyID)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID, colonyID, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(process2.ID, colonyID, "", core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteProcessByID(process1.ID)
	assert.Nil(t, err)

	_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.NotNil(t, err)

	attributeFromDB, err := db.GetAttribute(process2.ID, "test_key2", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Not deleted as it belongs to process 2
}

func TestAssign(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, processFromDB.State)
	assert.False(t, processFromDB.IsAssigned)

	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)

	err = db.Assign(executor.ID, process)
	assert.NotNil(t, err) // Should not work, already assigned

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.True(t, processFromDB.IsAssigned)
	assert.False(t, int64(processFromDB.StartTime.Sub(processFromDB.SubmissionTime)) < 0)
	assert.Equal(t, core.RUNNING, processFromDB.State)

	err = db.Unassign(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.False(t, processFromDB.IsAssigned)
	assert.False(t, int64(processFromDB.EndTime.Sub(processFromDB.StartTime)) < 0)
}

func TestMarkSuccessful(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, process.State)

	_, _, err = db.MarkSuccessful(process.ID)
	assert.NotNil(t, err) // Not possible to set waiting process to successfull

	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, process.State)

	_, _, err = db.MarkSuccessful(process.ID)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.SUCCESS, processFromDB.State)

	err = db.MarkFailed(process.ID, []string{"error"})
	assert.NotNil(t, err) // Not possible to set a successful process as failed

	process = utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)

	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	_, _, err = db.MarkSuccessful(process.ID)
	assert.NotNil(t, err) // Not possible to set a failed process to successful
}

func TestMarkFailed(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, process.State)

	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, processFromDB.State)

	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.Errors, []string{"error"})
	assert.Equal(t, core.FAILED, processFromDB.State)

	err = db.MarkFailed(process.ID, []string{"error"})
	assert.NotNil(t, err) // Not possible to set failed process as failed
}

func TestResetProcess(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.ID)
	process.FunctionSpec.MaxWaitTime = -1
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	process = utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	process = utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	numberOfFailedProcesses, err := db.CountFailedProcesses()
	assert.Equal(t, 3, numberOfFailedProcesses)

	err = db.ResetProcess(process)
	assert.Nil(t, err)

	numberOfFailedProcesses, err = db.CountFailedProcesses()
	assert.Equal(t, 2, numberOfFailedProcesses)
}

func TestSetWaitingForParents(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	err = db.SetWaitForParents(process.ID, true)
	assert.Nil(t, err)
	process2, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.True(t, process2.WaitForParents)

	err = db.SetWaitForParents(process.ID, false)
	assert.Nil(t, err)
	process2, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.False(t, process2.WaitForParents)
}

func TestSetParents(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	assert.Len(t, process.Parents, 0)

	parent := core.GenerateRandomID()
	parents := []string{parent}

	err = db.SetParents(process.ID, parents)
	assert.Nil(t, err)
	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Len(t, processFromDB.Parents, 1)
	assert.Equal(t, parent, processFromDB.Parents[0])
}

func TestSetChildren(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	assert.Len(t, process.Children, 0)

	child := core.GenerateRandomID()
	children := []string{child}

	err = db.SetChildren(process.ID, children)
	assert.Nil(t, err)
	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Len(t, processFromDB.Children, 1)
	assert.Equal(t, child, processFromDB.Children[0])
}

func TestSetProcessState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	err = db.SetProcessState(process.ID, core.RUNNING)
	assert.Nil(t, err)
	process2, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, process2.State, core.RUNNING)

	err = db.SetProcessState(process.ID, core.FAILED)
	assert.Nil(t, err)
	process2, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, process2.State, core.FAILED)
}

func TestSetWaitDeadline(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	assert.Equal(t, process.ExecDeadline, time.Time{})

	err = db.SetWaitDeadline(process, time.Now())
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.NotEqual(t, processFromDB.WaitDeadline, time.Time{})
}

func TestSetInput(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	input := make([]interface{}, 2)
	input[0] = "result1"
	input[1] = "result2"
	err = db.SetInput(process.ID, input)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Len(t, processFromDB.Input, 2)
	assert.Equal(t, processFromDB.Input[0], "result1")
	assert.Equal(t, processFromDB.Input[1], "result2")
}

func TestSetInput2(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	input := make([]interface{}, 2)
	input[0] = "result1"
	input[1] = "result2"
	process.Input = input
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Len(t, processFromDB.Input, 2)
	assert.Equal(t, processFromDB.Input[0], "result1")
	assert.Equal(t, processFromDB.Input[1], "result2")
}

func TestSetOutput(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	output := make([]interface{}, 2)
	output[0] = "result1"
	output[1] = "result2"
	err = db.SetOutput(process.ID, output)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Len(t, processFromDB.Output, 2)
	assert.Equal(t, processFromDB.Output[0], "result1")
	assert.Equal(t, processFromDB.Output[1], "result2")
}

func TestSetExecDeadline(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	assert.Equal(t, process.ExecDeadline, time.Time{})

	err = db.SetExecDeadline(process, time.Now())
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.NotEqual(t, processFromDB.ExecDeadline, time.Time{})
}

func TestSetErrorMsg(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	assert.Len(t, process.Errors, 0)

	err = db.SetErrors(process.ID, []string{"error"})
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Len(t, processFromDB.Errors, 1)
	assert.Equal(t, processFromDB.Errors[0], "error")
}

func TestFindUnassignedProcesses1(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colony.ID)
	process2.WaitForParents = true
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, executor.ID, executor.Type, 100)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
}

func TestFindUnassignedProcesses2(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcessWithTargets(colony.ID, []string{executor2.ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colony.ID, []string{executor2.ID})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, executor2.ID, executor2.Type, 2)
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

// Test that the order of targetExecutorIDs strings does not matter
func TestFindUnassignedProcesses3(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcessWithTargets(colony.ID, []string{executor1.ID, executor2.ID})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcessWithTargets(colony.ID, []string{executor1.ID, executor2.ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, executor1.ID, executor1.Type, 1)
	assert.Nil(t, err)

	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID, executor2.ID, executor2.Type, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)
}

// Test that executor type matching is working
func TestFindUnassignedProcesses4(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutorWithType(colony.ID, "test_executor_type_1")
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutorWithType(colony.ID, "test_executor_type_2")
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcessWithType(colony.ID, "test_executor_type_1")
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcessWithType(colony.ID, "test_executor_type_2")
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, executor1.ID, executor1.Type, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)

	processsFromDB, err = db.FindUnassignedProcesses(colony.ID, executor2.ID, executor2.Type, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process2.ID, processsFromDB[0].ID)
}

func TestFindUnassignedProcessesOldest(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colony.ID)
	process2.WaitForParents = true
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindUnassignedProcesses(colony.ID, executor.ID, executor.Type, 100)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, processsFromDB[0].ID, process1.ID)
}

func TestFindProcessAssigned(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	numberOfProcesses, err := db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	numberOfRunningProcesses, err := db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfRunningProcesses)

	numberOfSuccesfulProcesses, err := db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfSuccesfulProcesses)

	numberOfFailedProcesses, err := db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfFailedProcesses)

	processsFromDB1, err := db.FindUnassignedProcesses(colony.ID, executor.ID, executor.Type, 1)
	assert.Nil(t, err)
	assert.Equal(t, process1.ID, processsFromDB1[0].ID)
	assert.Len(t, processsFromDB1, 1)

	err = db.Assign(executor.ID, processsFromDB1[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfRunningProcesses)

	processsFromDB2, err := db.FindUnassignedProcesses(colony.ID, executor.ID, executor.Type, 1)
	assert.Nil(t, err)
	assert.Equal(t, process2.ID, processsFromDB2[0].ID)

	err = db.Assign(executor.ID, processsFromDB2[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfRunningProcesses)

	_, _, err = db.MarkSuccessful(processsFromDB1[0].ID)
	assert.Nil(t, err)

	err = db.MarkFailed(processsFromDB2[0].ID, []string{"error"})
	assert.Nil(t, err)

	numberOfSuccesfulProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfSuccesfulProcesses)

	numberOfFailedProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfFailedProcesses)
}

func TestFindWaitingProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	waitingProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		waitingProcessIDs[process.ID] = true
	}
	waitingProcessIDsFromDB, err := db.FindWaitingProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Create some running processes
	runningProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		runningProcessIDs[process.ID] = true
	}
	runningProcessIDsFromDB, err := db.FindRunningProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Create some successful processes
	successfulProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
		successfulProcessIDs[process.ID] = true
	}
	successfulProcessIDsFromDB, err := db.FindSuccessfulProcesses(colony.ID, 20)
	assert.Nil(t, err)

	// Create some successful processes
	failedProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process.ID, []string{"error"})
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

	numberOfProcesses, err := db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 40, numberOfProcesses)

	numberOfProcesses, err = db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)

	numberOfProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 10, numberOfProcesses)
}

func TestFindAllProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony2.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony1.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
	}
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony2.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
	}

	// Create some running processes
	for i := 0; i < 5; i++ {
		process := utils.CreateTestProcess(colony1.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
	}
	for i := 0; i < 5; i++ {
		process := utils.CreateTestProcess(colony2.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor2.ID, process)
		assert.Nil(t, err)
	}

	runningProcessIDsFromDB, err := db.FindAllRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, len(runningProcessIDsFromDB), 10)

	waitingProcessIDsFromDB, err := db.FindAllWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, len(waitingProcessIDsFromDB), 20)
}

func TestFindProcessesByExecutorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony2.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony1.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
	}
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony2.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
	}

	// Create some running processes
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony1.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}
	for i := 0; i < 20; i++ {
		process := utils.CreateTestProcess(colony2.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor2.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	process := utils.CreateTestProcess(colony1.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process)
	assert.Nil(t, err)
	_, _, err = db.MarkSuccessful(process.ID)
	assert.Nil(t, err)

	processesFromDB, err := db.FindProcessesByExecutorID(colony1.ID, executor1.ID, 60, core.SUCCESS) // last 60 seconds
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 11)

	processesFromDB, err = db.FindProcessesByExecutorID(colony1.ID, executor1.ID, 1, core.SUCCESS) // last second
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 1)

	processesFromDB, err = db.FindProcessesByExecutorID(colony2.ID, executor2.ID, 60, core.SUCCESS)
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 20)
}

func TestFindProcessesByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	for i := 0; i < 20; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
	}

	// Create some running processes
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.ID)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor2.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	process := utils.CreateTestProcess(colony.ID)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process)
	assert.Nil(t, err)
	_, _, err = db.MarkSuccessful(process.ID)
	assert.Nil(t, err)

	processesFromDB, err := db.FindProcessesByColonyID(colony.ID, 60, core.SUCCESS) // last 60 seconds
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 21)

	processesFromDB, err = db.FindProcessesByColonyID(colony.ID, 1, core.SUCCESS) // last second
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 1)
}
