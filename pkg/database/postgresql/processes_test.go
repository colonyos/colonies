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

	_, err = db.FindProcessesByColonyName("invalid_name", 60, core.SUCCESS)
	assert.NotNil(t, err)

	_, err = db.FindProcessesByExecutorID("invalid_id", "invalid_id", 60, core.SUCCESS)
	assert.NotNil(t, err)

	_, err = db.FindWaitingProcesses("invalid_id", "", "", "", 1)
	assert.NotNil(t, err)

	_, err = db.FindRunningProcesses("invalid_id", "", "", "", 1)
	assert.NotNil(t, err)

	_, err = db.FindAllRunningProcesses()
	assert.NotNil(t, err)

	_, err = db.FindAllWaitingProcesses()
	assert.NotNil(t, err)

	_, err = db.FindSuccessfulProcesses("invalid_id", "", "", "", 1)
	assert.NotNil(t, err)

	_, err = db.FindFailedProcesses("invalid_id", "", "", "", 1)
	assert.NotNil(t, err)

	_, err = db.FindCandidates("invalid_id", "invalid_type", 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
	assert.NotNil(t, err)

	err = db.RemoveProcessByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllProcesses()
	assert.NotNil(t, err)

	err = db.RemoveAllWaitingProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveAllRunningProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveAllSuccessfulProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveAllFailedProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveAllProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveAllProcessesByProcessGraphID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllProcessesInProcessGraphsByColonyName("invalid_name")
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
	_, err = db.CountWaitingProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)
	_, err = db.CountRunningProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)
	_, err = db.CountSuccessfulProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)
	_, err = db.CountFailedProcessesByColonyName("invalid_name")
	assert.NotNil(t, err)
}

func TestAddProcess(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executor1Name := core.GenerateRandomID()
	executor2Name := core.GenerateRandomID()

	process := utils.CreateTestProcessWithTargets(colonyName, []string{executor1Name, executor2Name})
	invalidKwArgs := make(map[string]interface{})
	invalidKwArgs["name"] = func() {
	}
	process.FunctionSpec.KwArgs = invalidKwArgs
	err = db.AddProcess(process)
	assert.NotNil(t, err)

	invalidArgs := make([]interface{}, 1)
	invalidArgs[0] = func() {
	}
	process.FunctionSpec.Args = invalidArgs
	err = db.AddProcess(process)
	assert.NotNil(t, err)

	process = utils.CreateTestProcessWithTargets(colonyName, []string{executor1Name, executor2Name})
	invalidInput := make([]interface{}, 1)
	invalidInput[0] = func() {
	}
	process.Input = invalidInput
	err = db.AddProcess(process)
	assert.NotNil(t, err)

	process = utils.CreateTestProcessWithTargets(colonyName, []string{executor1Name, executor2Name})
	invalidOutput := make([]interface{}, 1)
	invalidOutput[0] = func() {
	}
	process.Output = invalidOutput
	err = db.AddProcess(process)
	assert.NotNil(t, err)

	process = utils.CreateTestProcessWithTargets(colonyName, []string{executor1Name, executor2Name})
	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Contains(t, processFromDB.FunctionSpec.Conditions.ExecutorNames, executor1Name)
	assert.Contains(t, processFromDB.FunctionSpec.Conditions.ExecutorNames, executor2Name)

	process = utils.CreateTestProcessWithTargets(colonyName, []string{executor1Name, executor2Name})

	var snapshots []core.SnapshotMount
	snapshot1 := core.SnapshotMount{Label: "test_label1", SnapshotID: "test_snapshotid1", Dir: "test_dir1", KeepFiles: false, KeepSnaphot: false}
	snapshot2 := core.SnapshotMount{Label: "test_label2", SnapshotID: "test_snapshotid2", Dir: "test_dir2", KeepFiles: true, KeepSnaphot: true}
	snapshots = append(snapshots, snapshot1)
	snapshots = append(snapshots, snapshot2)
	var syncdirs []core.SyncDirMount
	syncdir1 := core.SyncDirMount{Label: "test_label1", Dir: "test_dir1", KeepFiles: false}
	syncdir2 := core.SyncDirMount{Label: "test_label2", Dir: "test_dir2", KeepFiles: false}
	syncdirs = append(syncdirs, syncdir1)
	syncdirs = append(syncdirs, syncdir2)
	process.FunctionSpec.Filesystem = core.Filesystem{SnapshotMounts: snapshots, SyncDirMounts: syncdirs, Mount: "/cfs"}

	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	assert.Len(t, processFromDB.FunctionSpec.Filesystem.SnapshotMounts, 2)
	assert.Len(t, processFromDB.FunctionSpec.Filesystem.SyncDirMounts, 2)
	assert.Equal(t, processFromDB.FunctionSpec.Filesystem.Mount, "/cfs")
	assert.Equal(t, processFromDB.FunctionSpec.Filesystem.SnapshotMounts[0].Label, "test_label1")
}

func TestAddProcessConditions(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	process := utils.CreateTestProcess(colonyName)
	process.FunctionSpec.Conditions.Nodes = 1
	process.FunctionSpec.Conditions.Processes = 2
	process.FunctionSpec.Conditions.ProcessesPerNode = 1
	process.FunctionSpec.Conditions.CPU = "1000m"
	process.FunctionSpec.Conditions.Memory = "10G"
	process.FunctionSpec.Conditions.Storage = "2000G"
	process.FunctionSpec.Conditions.WallTime = 70
	process.FunctionSpec.Conditions.GPU.Name = "nvidia_2080ti"
	process.FunctionSpec.Conditions.GPU.Count = 4
	process.FunctionSpec.Conditions.GPU.Memory = "10G"

	err = db.AddProcess(process)
	assert.Nil(t, err)

	processFromDB, err := db.GetProcessByID(process.ID)
	assert.Nil(t, err)
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.Nodes, 1)
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.Processes, 2)
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.ProcessesPerNode, 1)
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.CPU, "1000m")
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.Memory, "10000000Mi")
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.Storage, "2000000000Mi")
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.WallTime, int64(70))
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.GPU.Name, "nvidia_2080ti")
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.GPU.Count, 4)
	assert.Equal(t, processFromDB.FunctionSpec.Conditions.GPU.Memory, "10000000Mi")
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

	colonyName := core.GenerateRandomID()
	process := utils.CreateTestProcessWithEnv(colonyName, env)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	_, err = db.GetProcessByID(process.ID)
	assert.Nil(t, err)

	processesFromDB, err := db.GetProcesses()
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 1)
}

func TestRemoveProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process1 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colonyName, []string{executor1ID, executor2ID})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	numberOfProcesses, err := db.CountWaitingProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	numberOfProcesses, err = db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfProcesses)

	err = db.RemoveProcessByID(process1.ID)
	assert.Nil(t, err)

	numberOfProcesses, err = db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfProcesses)

	err = db.RemoveAllProcesses()
	assert.Nil(t, err)

	numberOfProcesses, err = db.CountProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfProcesses)
}

func TestRemoveAllProcessesByColony(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1Name := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colony1Name)
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colony1Name, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	colony2Name := core.GenerateRandomID()
	process2 := utils.CreateTestProcess(colony2Name)
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colony2Name, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	err = db.RemoveAllProcessesByColonyName(colony2Name)
	assert.Nil(t, err)

	_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.Nil(t, err)
	_, err = db.GetAttribute(process2.ID, "test_key1", core.IN)
	assert.NotNil(t, err)
}

func TestRemoveAllProcessesByColonyWithState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1Name := core.GenerateRandomID()
	colony2Name := core.GenerateRandomID()
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()

	process1 := utils.CreateTestProcessWithTargets(colony1Name, []string{executor1ID, executor2ID})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithTargets(colony1Name, []string{executor1ID, executor2ID})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colony1Name, []string{executor1ID, executor2ID})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithTargets(colony1Name, []string{executor1ID, executor2ID})
	err = db.AddProcess(process4)
	assert.Nil(t, err)

	process5 := utils.CreateTestProcessWithTargets(colony2Name, []string{executor1ID, executor2ID})
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

	err = db.RemoveAllWaitingProcessesByColonyName(colony1Name)
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

	err = db.RemoveAllRunningProcessesByColonyName(colony1Name)
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

	err = db.RemoveAllSuccessfulProcessesByColonyName(colony1Name)
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

	err = db.RemoveAllFailedProcessesByColonyName(colony1Name)
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

	err = db.RemoveAllFailedProcessesByColonyName(colony2Name)
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

func TestRemoveAllProcessesByProcessGraphID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	processGraphID := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyName)
	process1.ProcessGraphID = processGraphID
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colonyName, processGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyName)
	process2.ProcessGraphID = processGraphID
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colonyName, processGraphID, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	attribute3 := core.CreateAttribute(process3.ID, colonyName, "", core.IN, "test_key1", "test_value1")
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

	err = db.RemoveAllProcessesByProcessGraphID(processGraphID)
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

func TestRemoveAllProcessesInProcessGraphsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyName)
	process1.ProcessGraphID = processGraphID1
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colonyName, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyName)
	process2.ProcessGraphID = processGraphID2
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colonyName, processGraphID2, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	attribute3 := core.CreateAttribute(process3.ID, colonyName, "", core.IN, "test_key1", "test_value1")
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

	err = db.RemoveAllProcessesInProcessGraphsByColonyName(colonyName)
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

func TestRemoveAllProcessesInProcessGraphsByColonyNameWithState(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	processGraphID1 := core.GenerateRandomID()
	processGraphID2 := core.GenerateRandomID()
	processGraphID3 := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyName)
	process1.ProcessGraphID = processGraphID1
	err = db.AddProcess(process1)
	assert.Nil(t, err)
	attribute1 := core.CreateAttribute(process1.ID, colonyName, processGraphID1, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyName)
	process2.ProcessGraphID = processGraphID2
	err = db.AddProcess(process2)
	assert.Nil(t, err)
	attribute2 := core.CreateAttribute(process2.ID, colonyName, processGraphID2, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colonyName)
	process3.ProcessGraphID = processGraphID3
	err = db.AddProcess(process3)
	assert.Nil(t, err)
	attribute3 := core.CreateAttribute(process3.ID, colonyName, processGraphID3, core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process4)
	assert.Nil(t, err)
	attribute4 := core.CreateAttribute(process4.ID, colonyName, "", core.IN, "test_key1", "test_value1")
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

	err = db.RemoveAllProcessesInProcessGraphsByColonyNameWithState(colonyName, core.FAILED)
	assert.Nil(t, err)

	runningProcesses, err = db.CountFailedProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 0)

	err = db.RemoveAllProcessesInProcessGraphsByColonyNameWithState(colonyName, core.RUNNING)
	assert.Nil(t, err)

	runningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, runningProcesses, 1)
}

func TestRemoveAllProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID, colonyName, core.GenerateRandomID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.RemoveAllProcesses()
	assert.Nil(t, err)

	_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.NotNil(t, err)
}

func TestRemoveProcessesAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	process1 := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process1.ID, colonyName, "", core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(process2.ID, colonyName, "", core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.RemoveProcessByID(process1.ID)
	assert.Nil(t, err)

	_, err = db.GetAttribute(process1.ID, "test_key1", core.IN)
	assert.NotNil(t, err)

	attributeFromDB, err := db.GetAttribute(process2.ID, "test_key2", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Not removed as it belongs to process 2
}

func TestAssign(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.Name)
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

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.Name)
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

	process = utils.CreateTestProcess(colony.Name)
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

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.Name)
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

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process := utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.MaxWaitTime = -1
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	process = utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor.ID, process)
	assert.Nil(t, err)
	err = db.MarkFailed(process.ID, []string{"error"})
	assert.Nil(t, err)

	process = utils.CreateTestProcess(colony.Name)
	process.FunctionSpec.MaxWaitTime = -1
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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
	process := utils.CreateTestProcess(colony.Name)
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

func TestFindCandidates1(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.WaitForParents = true
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindCandidates(colony.Name, executor.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 100)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
}

func TestFindCandidates2(t *testing.T) {
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

	process1 := utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcessWithTargets(colony.Name, []string{executor2.Name})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithTargets(colony.Name, []string{executor2.Name})
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processesFromDB, err := db.FindCandidates(colony.Name, executor2.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 2)
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 1)
	assert.Equal(t, processesFromDB[0].ID, process1.ID)

	processesFromDB, err = db.FindCandidatesByName(colony.Name, executor2.Name, executor2.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 2)
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 2)

	counter := 0
	for _, processFromDB := range processesFromDB {
		if processFromDB.ID == process2.ID {
			counter++
		}

		if processFromDB.ID == process3.ID {
			counter++
		}
	}

	assert.Equal(t, 2, counter)
}

// Test that the order of targetExecutorIDs strings does not matter
func TestFindCandidates3(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcessWithTargets(colony.Name, []string{executor1.Name, executor2.Name})
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcessWithTargets(colony.Name, []string{executor1.Name, executor2.Name})
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processesFromDB, err := db.FindCandidates(colony.Name, executor1.Type, 0, 0, "", 0, 0, 0, 0, 0, 9, 1)
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 0)

	processesFromDB, err = db.FindCandidatesByName(colony.Name, executor1.Name, executor1.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 1)
	assert.Equal(t, processesFromDB[0].ID, process1.ID)

	processesFromDB, err = db.FindCandidatesByName(colony.Name, executor2.Name, executor1.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
	assert.Nil(t, err)
	assert.Len(t, processesFromDB, 1)
	assert.Equal(t, processesFromDB[0].ID, process1.ID)
}

// Test that executor type matching is working
func TestFindCandidates4(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutorWithType(colony.Name, "test_executor_type_1")
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutorWithType(colony.Name, "test_executor_type_2")
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcessWithType(colony.Name, "test_executor_type_1")
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcessWithType(colony.Name, "test_executor_type_2")
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindCandidates(colony.Name, executor1.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process1.ID, processsFromDB[0].ID)

	processsFromDB, err = db.FindCandidates(colony.Name, executor2.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, process2.ID, processsFromDB[0].ID)
}

func TestFindCandidatesOldest(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.WaitForParents = true
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	processsFromDB, err := db.FindCandidates(colony.Name, executor.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 100)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)
	assert.Equal(t, processsFromDB[0].ID, process1.ID)
}

func TestFindCandidatesByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	executor1.Name = "executor1"
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	executor2.Name = "executor2"
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcess(colony.Name)
	process2.FunctionSpec.Conditions.ExecutorNames = []string{"executor1"}
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcess(colony.Name)
	process3.FunctionSpec.Conditions.ExecutorNames = []string{"executor1", "executor2"}
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	processsFromDB, err := db.FindCandidates(colony.Name, executor1.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 100)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 1)

	processsFromDB, err = db.FindCandidatesByName(colony.Name, "executor1", executor1.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 100)
	assert.Nil(t, err)
	assert.Len(t, processsFromDB, 2)

	counter := 0
	for _, process := range processsFromDB {
		if process.ID == process2.ID {
			counter++
		}
		if process.ID == process3.ID {
			counter++
		}
	}

	assert.True(t, counter == 2)
}

func TestFindProcessAssigned(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	process1 := utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := utils.CreateTestProcess(colony.Name)
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

	processsFromDB1, err := db.FindCandidates(colony.Name, executor.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
	assert.Nil(t, err)
	assert.Equal(t, process1.ID, processsFromDB1[0].ID)
	assert.Len(t, processsFromDB1, 1)

	err = db.Assign(executor.ID, processsFromDB1[0])
	assert.Nil(t, err)

	numberOfRunningProcesses, err = db.CountRunningProcesses()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfRunningProcesses)

	processsFromDB2, err := db.FindCandidates(colony.Name, executor.Type, 0, 0, "", 0, 0, 0, 0, 0, 0, 1)
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

func TestFindProcesses(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.Name)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	waitingProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		waitingProcessIDs[process.ID] = true
	}
	waitingProcessIDsFromDB, err := db.FindWaitingProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)

	// Create some running processes
	runningProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		runningProcessIDs[process.ID] = true
	}
	runningProcessIDsFromDB, err := db.FindRunningProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)

	// Create some successful processes
	successfulProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
		successfulProcessIDs[process.ID] = true
	}
	successfulProcessIDsFromDB, err := db.FindSuccessfulProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)

	// Create some failed processes
	failedProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor.ID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process.ID, []string{"error"})
		assert.Nil(t, err)
		failedProcessIDs[process.ID] = true
	}
	failedProcessIDsFromDB, err := db.FindFailedProcesses(colony.Name, "", "", "", 20)
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

func TestFindProcessesByFilter(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.Name)
	executor1.Type = "test_executor_type_1"
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.Name)
	executor2.Type = "test_executor_type_2"
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	// Create some waiting/unassigned processes
	waitingProcessIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_1"
		process.FunctionSpec.Label = "test_label_1"
		err = db.AddProcess(process)
		assert.Nil(t, err)
		waitingProcessIDs[process.ID] = true
	}
	for i := 0; i < 5; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_2"
		process.FunctionSpec.Label = "test_label_1"
		err = db.AddProcess(process)
		assert.Nil(t, err)
		waitingProcessIDs[process.ID] = true
	}
	waitingProcessIDsFromDB, err := db.FindWaitingProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, waitingProcessIDsFromDB, 15)

	waitingProcessIDsFromDB, err = db.FindWaitingProcesses(colony.Name, "test_executor_type_1", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, waitingProcessIDsFromDB, 10)

	waitingProcessIDsFromDB, err = db.FindWaitingProcesses(colony.Name, "test_executor_type_2", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, waitingProcessIDsFromDB, 5)

	// Create some running processes
	for i := 0; i < 4; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_1"
		process.FunctionSpec.Label = "test_label_1"
		err = db.AddProcess(process)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
	}
	for i := 0; i < 3; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_2"
		err = db.AddProcess(process)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
	}

	runningProcessIDsFromDB, err := db.FindRunningProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, runningProcessIDsFromDB, 7)

	runningProcessIDsFromDB, err = db.FindRunningProcesses(colony.Name, "test_executor_type_1", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, runningProcessIDsFromDB, 4)

	runningProcessIDsFromDB, err = db.FindRunningProcesses(colony.Name, "test_executor_type_2", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, runningProcessIDsFromDB, 3)

	// Create some successful processes
	for i := 0; i < 6; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_1"
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}
	for i := 0; i < 12; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_2"
		process.FunctionSpec.Label = "test_label_1"
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}
	successfulProcessIDsFromDB, err := db.FindSuccessfulProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, successfulProcessIDsFromDB, 18)

	successfulProcessIDsFromDB, err = db.FindSuccessfulProcesses(colony.Name, "test_executor_type_1", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, successfulProcessIDsFromDB, 6)

	successfulProcessIDsFromDB, err = db.FindSuccessfulProcesses(colony.Name, "test_executor_type_2", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, successfulProcessIDsFromDB, 12)

	// Create some failed processes
	for i := 0; i < 3; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_1"
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process.ID, []string{"error"})
		assert.Nil(t, err)
	}
	for i := 0; i < 2; i++ {
		process := utils.CreateTestProcess(colony.Name)
		process.InitiatorName = "test_initiator_name_1"
		process.FunctionSpec.Conditions.ExecutorType = "test_executor_type_2"
		process.FunctionSpec.Label = "test_label_1"
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		err = db.MarkFailed(process.ID, []string{"error"})
		assert.Nil(t, err)
	}
	failedProcessIDsFromDB, err := db.FindFailedProcesses(colony.Name, "", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, failedProcessIDsFromDB, 5)

	failedProcessIDsFromDB, err = db.FindFailedProcesses(colony.Name, "test_executor_type_1", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, failedProcessIDsFromDB, 3)

	failedProcessIDsFromDB, err = db.FindFailedProcesses(colony.Name, "test_executor_type_2", "", "", 20)
	assert.Nil(t, err)
	assert.Len(t, failedProcessIDsFromDB, 2)

	// Label filter
	p, err := db.FindWaitingProcesses(colony.Name, "", "test_label_1", "", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 15)
	for _, process := range p {
		assert.Equal(t, process.FunctionSpec.Label, "test_label_1")
	}

	p, err = db.FindRunningProcesses(colony.Name, "", "test_label_1", "", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 4)
	for _, process := range p {
		assert.Equal(t, process.FunctionSpec.Label, "test_label_1")
	}

	p, err = db.FindSuccessfulProcesses(colony.Name, "", "test_label_1", "", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 12)
	for _, process := range p {
		assert.Equal(t, process.FunctionSpec.Label, "test_label_1")
	}

	p, err = db.FindFailedProcesses(colony.Name, "", "test_label_1", "", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 2)
	for _, process := range p {
		assert.Equal(t, process.FunctionSpec.Label, "test_label_1")
	}

	// Initiator filter
	p, err = db.FindWaitingProcesses(colony.Name, "", "", "test_initiator_name_1", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 15)
	for _, process := range p {
		assert.Equal(t, process.InitiatorName, "test_initiator_name_1")
	}

	p, err = db.FindRunningProcesses(colony.Name, "", "", "test_initiator_name_1", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 7)
	for _, process := range p {
		assert.Equal(t, process.InitiatorName, "test_initiator_name_1")
	}

	p, err = db.FindSuccessfulProcesses(colony.Name, "", "", "test_initiator_name_1", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 18)
	for _, process := range p {
		assert.Equal(t, process.InitiatorName, "test_initiator_name_1")
	}

	p, err = db.FindFailedProcesses(colony.Name, "", "", "test_initiator_name_1", 20)
	assert.Nil(t, err)
	assert.Len(t, p, 5)
	for _, process := range p {
		assert.Equal(t, process.InitiatorName, "test_initiator_name_1")
	}
}

func TestFindAllProcesses(t *testing.T) {
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

	executor2 := utils.CreateTestExecutor(colony2.Name)
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

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.Name)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony2.Name)
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

func TestFindProcessesByColonyName(t *testing.T) {
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

	// Create some waiting/unassigned processes
	for i := 0; i < 20; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
	}

	// Create some running processes
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor1.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}
	for i := 0; i < 10; i++ {
		process := utils.CreateTestProcess(colony.Name)
		err = db.AddProcess(process)
		assert.Nil(t, err)
		err = db.Assign(executor2.ID, process)
		assert.Nil(t, err)
		_, _, err = db.MarkSuccessful(process.ID)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	process := utils.CreateTestProcess(colony.Name)
	err = db.AddProcess(process)
	assert.Nil(t, err)
	err = db.Assign(executor1.ID, process)
	assert.Nil(t, err)
	_, _, err = db.MarkSuccessful(process.ID)
	assert.Nil(t, err)

	processesFromDB, err := db.FindProcessesByColonyName(colony.Name, 60, core.SUCCESS) // last 60 seconds
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 21)

	processesFromDB, err = db.FindProcessesByColonyName(colony.Name, 1, core.SUCCESS) // last second
	assert.Nil(t, err)
	assert.Equal(t, len(processesFromDB), 1)
}
