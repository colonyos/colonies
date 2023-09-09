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

	err = db.AddOrReplaceExecutor(executor)
	assert.NotNil(t, err)

	_, err = db.GetExecutors()
	assert.NotNil(t, err)

	_, err = db.GetExecutorByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetExecutorsByColonyID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetExecutorByName("invalid_id", "invalid_name")
	assert.NotNil(t, err)

	err = db.ApproveExecutor(executor)
	assert.NotNil(t, err)

	err = db.RejectExecutor(executor)
	assert.NotNil(t, err)

	err = db.MarkAlive(executor)
	assert.NotNil(t, err)

	err = db.DeleteExecutorByID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteExecutorsByColonyID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.CountExecutors()
	assert.NotNil(t, err)

	_, err = db.CountExecutorsByColonyID("invalid_id")
	assert.NotNil(t, err)
}

func TestAddExecutor(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	executor.Capabilities.Software.Name = "sw_name"
	executor.Capabilities.Software.Type = "sw_type"
	executor.Capabilities.Software.Version = "sw_version"

	executor.Capabilities.Hardware.Model = "model"
	executor.Capabilities.Hardware.Nodes = 10
	executor.Capabilities.Hardware.CPU = "1000m"
	executor.Capabilities.Hardware.Memory = "10G"
	executor.Capabilities.Hardware.Storage = "1000G"
	executor.Capabilities.Hardware.GPU.Name = "nvidia_2080ti"
	executor.Capabilities.Hardware.GPU.Count = 4000
	executor.Capabilities.Hardware.GPU.NodeCount = 4
	executor.Capabilities.Hardware.GPU.Memory = "10G"

	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executors, err := db.GetExecutors()
	assert.Nil(t, err)

	executorFromDB := executors[0]
	assert.True(t, executor.Equals(executorFromDB))
	assert.True(t, executorFromDB.IsPending())
	assert.False(t, executorFromDB.IsApproved())
	assert.False(t, executorFromDB.IsRejected())

	assert.Equal(t, executor.Capabilities.Software.Name, "sw_name")
	assert.Equal(t, executor.Capabilities.Software.Type, "sw_type")
	assert.Equal(t, executor.Capabilities.Software.Version, "sw_version")

	assert.Equal(t, executor.Capabilities.Hardware.Model, "model")
	assert.Equal(t, executor.Capabilities.Hardware.Nodes, 10)
	assert.Equal(t, executor.Capabilities.Hardware.CPU, "1000m")
	assert.Equal(t, executor.Capabilities.Hardware.Memory, "10G")
	assert.Equal(t, executor.Capabilities.Hardware.Storage, "1000G")
	assert.Equal(t, executor.Capabilities.Hardware.GPU.Name, "nvidia_2080ti")
	assert.Equal(t, executor.Capabilities.Hardware.GPU.Count, 4000)
	assert.Equal(t, executor.Capabilities.Hardware.GPU.NodeCount, 4)
	assert.Equal(t, executor.Capabilities.Hardware.GPU.Memory, "10G")
}

func TestAddOrReplaceExecutor(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony.ID)
	executor.Name = "test_name_1"
	err = db.AddOrReplaceExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.Equal(t, executorFromDB.Name, "test_name_1")

	executor.Name = "test_name_2"
	err = db.AddOrReplaceExecutor(executor)
	assert.Nil(t, err)

	executorFromDB, err = db.GetExecutorByID(executor.ID)
	assert.Nil(t, err)
	assert.Equal(t, executorFromDB.Name, "test_name_2")
}

func TestAddTwoExecutors(t *testing.T) {
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

	var executors []*core.Executor
	executors = append(executors, executor1)
	executors = append(executors, executor2)

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

	executor1 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID("invalid_id")
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByID(executor1.ID)
	assert.Nil(t, err)
	assert.True(t, executor1.Equals(executorFromDB))
}

func TestGetExecutorByColonyID(t *testing.T) {
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

	executor1 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.ID)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	var executorsColony1 []*core.Executor
	executorsColony1 = append(executorsColony1, executor1)
	executorsColony1 = append(executorsColony1, executor2)

	executorsColony1FromDB, err := db.GetExecutorsByColonyID("invalid_id")
	assert.Nil(t, err)
	assert.NotNil(t, executorsColony1)

	executorsColony1FromDB, err = db.GetExecutorsByColonyID(colony1.ID)
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

	executor1 := utils.CreateTestExecutor(colony.ID)
	executor1.Name = "test_name_1"
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.ID)
	executor2.Name = "test_name_"
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByName("invalid__id", executor1.Name)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByName(colony.ID, "invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByName("invalid__id", "invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	executorFromDB, err = db.GetExecutorByName(colony.ID, executor1.Name)
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

	executor := utils.CreateTestExecutor(colony.ID)
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

	executor := utils.CreateTestExecutor(colony.ID)
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

func TestDeleteExecutorMoveBackToQueue(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor1.ID, ColonyID: colony.ID, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor2.ID, ColonyID: colony.ID, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	env := make(map[string]string)

	process1 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithEnv(colony.ID, env)
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

	count, err := db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 0)

	_, _, err = db.MarkSuccessful(process4.ID)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colony.ID)
	assert.Len(t, functions, 2)

	err = db.DeleteExecutorByID(executor1.ID)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyID(colony.ID)
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

	count, err = db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 2)

	count, err = db.CountSuccessfulProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 1)

	count, err = db.CountRunningProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 1)

	count, err = db.CountFailedProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 0)
}

func TestDeleteExecutorsMoveBackToQueue(t *testing.T) {
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

	env := make(map[string]string)

	process1 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process1)
	assert.Nil(t, err)

	process2 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process2)
	assert.Nil(t, err)

	process3 := utils.CreateTestProcessWithEnv(colony.ID, env)
	err = db.AddProcess(process3)
	assert.Nil(t, err)

	process4 := utils.CreateTestProcessWithEnv(colony.ID, env)
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

	count, err := db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 0)

	_, _, err = db.MarkSuccessful(process4.ID)
	assert.Nil(t, err)

	err = db.DeleteExecutorsByColonyID(colony.ID)
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

	count, err = db.CountWaitingProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 3)

	count, err = db.CountSuccessfulProcessesByColonyID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, count == 1)
}

func TestDeleteExecutors(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor1 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor1)
	assert.Nil(t, err)

	function := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor1.ID, ColonyID: colony1.ID, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor2 := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor2.ID, ColonyID: colony1.ID, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	executor3 := utils.CreateTestExecutor(colony2.ID)
	err = db.AddExecutor(executor3)
	assert.Nil(t, err)

	function = &core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executor3.ID, ColonyID: colony2.ID, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	err = db.AddFunction(function)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colony1.ID)
	assert.Len(t, functions, 2)

	functions, err = db.GetFunctionsByColonyID(colony2.ID)
	assert.Len(t, functions, 1)

	err = db.DeleteExecutorByID(executor2.ID)
	assert.Nil(t, err)

	executorFromDB, err := db.GetExecutorByID(executor2.ID)
	assert.Nil(t, err)
	assert.Nil(t, executorFromDB)

	err = db.AddExecutor(executor2)
	assert.Nil(t, err)

	executorFromDB, err = db.GetExecutorByID(executor2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, executorFromDB)

	err = db.DeleteExecutorsByColonyID(colony1.ID)
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

	functions, err = db.GetFunctionsByColonyID(colony1.ID)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyID(colony2.ID)
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

	executor := utils.CreateTestExecutor(colony.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorCount, err = db.CountExecutors()
	assert.Nil(t, err)
	assert.True(t, executorCount == 1)
}

func TestCountExectorsByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	executor := utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executor = utils.CreateTestExecutor(colony1.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	executor = utils.CreateTestExecutor(colony2.ID)
	err = db.AddExecutor(executor)
	assert.Nil(t, err)

	executorCount, err := db.CountExecutors()
	assert.Nil(t, err)
	assert.True(t, executorCount == 3)

	executorCount, err = db.CountExecutorsByColonyID(colony1.ID)
	assert.Nil(t, err)
	assert.True(t, executorCount == 2)

	executorCount, err = db.CountExecutorsByColonyID(colony2.ID)
	assert.Nil(t, err)
	assert.True(t, executorCount == 1)

}
