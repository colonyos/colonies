package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubmitProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	funcSpec1 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, in)
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, funcSpec2.Conditions.ColonyID, addedProcess2.FunctionSpec.Conditions.ColonyID)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.colonyID, 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArraysEqual(processes, processesFromServer))

	server.Shutdown()
	<-done
}

func TestSubmitProcessInvalidPriority(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	funcSpec := utils.CreateTestFunctionSpecWithEnv(env.colonyID, in)
	funcSpec.Priority = MIN_PRIORITY - 100
	_, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.NotNil(t, err)

	funcSpec.Priority = MAX_PRIORITY + 100
	_, err = client.Submit(funcSpec, env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, make(map[string]string))
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessWithTimeout(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	addedProcessChan := make(chan *core.Process)
	go func() {
		time.Sleep(1 * time.Second)
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		addedProcessChan <- addedProcess
	}()

	// This function call will block for 60 seconds or until the Go-routine above submits a process spec
	assignProcess, err := client.Assign(env.colonyID, 60, env.executorPrvKey)
	assert.Nil(t, err)

	addedProcess := <-addedProcessChan
	assert.Equal(t, addedProcess.ID, assignProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessWithTimeoutFail(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	_, err := client.Assign(env.colonyID, 1, env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestAssignProcessNoPriority(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	funcSpec1.Priority = 0
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, make(map[string]string))
	funcSpec2.Priority = 0
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, make(map[string]string))
	funcSpec3.Priority = 0
	addedProcess3, err := client.Submit(funcSpec3, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessPriority(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	funcSpec1.Priority = 1
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, make(map[string]string))
	funcSpec2.Priority = 2
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, make(map[string]string))
	funcSpec3.Priority = 5
	addedProcess3, err := client.Submit(funcSpec3, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.colonyID, make(map[string]string))
	funcSpec4.Priority = 5
	addedProcess4, err := client.Submit(funcSpec4, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestMarkAlive(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(executor.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	executorFromServer, err := client.GetExecutor(executor.ID, executorPrvKey)
	assert.Nil(t, err)

	time1 := executorFromServer.LastHeardFromTime
	time.Sleep(1 * time.Second)

	client.Assign(env.colonyID, -1, executorPrvKey) // This will update the last heard from

	executorFromServer, err = client.GetExecutor(executor.ID, executorPrvKey)
	assert.Nil(t, err)
	time2 := executorFromServer.LastHeardFromTime

	assert.True(t, time1 != time2)

	server.Shutdown()
	<-done
}

func TestGetProcessHistForColony(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 3
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}

	// Get processes for the last 60 seconds
	processesFromServer, err := client.GetProcessHistForColony(core.WAITING, env.colonyID, 60, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	server.Shutdown()
	<-done
}

func TestGetProcessHistForExecutor(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	numberOfRunningProcesses := 10
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colony1ID)
		_, err := client.Submit(funcSpec, env.executor1PrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.colony1ID, -1, env.executor1PrvKey)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	funcSpec := utils.CreateTestFunctionSpec(env.colony1ID)
	_, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	_, err = client.Assign(env.colony1ID, -1, env.executor1PrvKey)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	// Get processes for the 60 seconds
	processesFromServer, err := client.GetProcessHistForExecutor(core.RUNNING, env.colony1ID, env.executor1ID, 60, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses+1)

	// Get processes for the last 2 seconds
	processesFromServer, err = client.GetProcessHistForExecutor(core.RUNNING, env.colony1ID, env.executor1ID, 2, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 1)

	server.Shutdown()
	<-done
}

func TestGetWaitingProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetWaitingProcesses(env.colonyID, numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetWaitingProcesses(env.colonyID, 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetRunningProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetRunningProcesses(env.colonyID, numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetRunningProcesses(env.colonyID, 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetSuccessfulProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		err = client.Close(processFromServer.ID, env.executorPrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetSuccessfulProcesses(env.colonyID, numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetSuccessfulProcesses(env.colonyID, 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetFailedProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		err = client.Fail(processFromServer.ID, []string{"error"}, env.executorPrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetFailedProcesses(env.colonyID, numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetFailedProcesses(env.colonyID, 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.executorPrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestDeleteProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.executorPrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	err = client.DeleteProcess(addedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err = client.GetProcess(addedProcess.ID, env.executorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, processFromServer)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessesForColony(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colony1ID)
	addedProcess1, err := client.Submit(funcSpec1, env.executor1PrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colony2ID)
	addedProcess2, err := client.Submit(funcSpec2, env.executor2PrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess1.ID, env.executor1PrvKey)
	assert.True(t, addedProcess1.Equals(processFromServer))

	err = client.DeleteAllProcesses(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcess(addedProcess1.ID, env.executor1PrvKey)
	assert.NotNil(t, err)

	processFromServer, err = client.GetProcess(addedProcess2.ID, env.executor2PrvKey)
	assert.Nil(t, err)
	assert.True(t, addedProcess2.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessesForColonyWithStateWaiting(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 2)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.DeleteAllProcessesWithState(env.colonyID, core.WAITING, env.colonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 0)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessesForColonyWithStateRunning(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 1)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.DeleteAllProcessesWithState(env.colonyID, core.RUNNING, env.colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessesForColonyWithStateSuccessful(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	err = client.Close(process.ID, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 1)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.DeleteAllProcessesWithState(env.colonyID, core.SUCCESS, env.colonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessesForColonyWithStateFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyID)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	err = client.Fail(process.ID, []string{"error"}, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 1)

	err = client.DeleteAllProcessesWithState(env.colonyID, core.FAILED, env.colonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	server.Shutdown()
	<-done
}

func TestCloseSuccessful(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Close(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.State)

	server.Shutdown()
	<-done
}

func TestCloseSuccessfulWithFunctions(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	_, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	function := &core.Function{ColonyID: env.colonyID, ExecutorID: env.executorID, FuncName: funcSpec.FuncName, Args: []string{}}
	_, err = client.AddFunction(function, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Close(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.State)

	functions, err := client.GetFunctionsByExecutorID(env.executorID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, functions[0].Counter, 1)
	assert.Greater(t, functions[0].MinWaitTime, 0.0)
	assert.Greater(t, functions[0].MaxWaitTime, 0.0)
	assert.Greater(t, functions[0].MinExecTime, 0.0)
	assert.Greater(t, functions[0].MaxExecTime, 0.0)
	assert.Greater(t, functions[0].AvgWaitTime, 0.0)
	assert.Greater(t, functions[0].AvgExecTime, 0.0)

	server.Shutdown()
	<-done
}

func TestCloseSuccessfulWithOutput(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	_, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	output := make([]interface{}, 2)
	output[0] = "result1"
	output[1] = "result2"
	err = client.CloseWithOutput(assignedProcess.ID, output, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	assert.Len(t, processFromServer.Output, 2)
	assert.Equal(t, processFromServer.Output[0], "result1")
	assert.Equal(t, processFromServer.Output[1], "result2")

	server.Shutdown()
	<-done
}

func TestCloseFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Fail(assignedProcess.ID, []string{"error"}, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Equal(t, processFromServer.State, core.FAILED)
	assert.Len(t, processFromServer.Errors, 1)
	assert.Equal(t, processFromServer.Errors[0], "error")

	server.Shutdown()
	<-done
}

func TestMaxWaitTime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	funcSpec.MaxWaitTime = 1 // 1 second

	process, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	var processes []*core.Process
	processes = append(processes, process)
	waitForProcesses(t, server, processes, core.FAILED)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, 1)

	server.Shutdown()
	<-done
}

func TestMaxExecTime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	funcSpec.MaxExecTime = 1 // 1 second

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	waitForProcesses(t, server, processes, core.WAITING)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeUnlimtedMaxRetries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	funcSpec.MaxExecTime = 1 // 1 second
	funcSpec.MaxRetries = -1 // Unlimted number of retries

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	waitForProcesses(t, server, processes, core.WAITING)

	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	// Assign again
	for i := 0; i < numberOfProcesses; i++ {
		_, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
	}

	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, numberOfProcesses)

	waitForProcesses(t, server, processes, core.WAITING)

	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeMaxRetries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyID)
	funcSpec.MaxExecTime = 3 // 3 seconds
	funcSpec.MaxRetries = 1  // Max 1 retries

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	waitForProcesses(t, server, processes, core.WAITING)

	// We should now have 10 waiting processes
	stat, err := client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	// Assign again
	for i := 0; i < numberOfProcesses; i++ {
		_, err = client.Assign(env.colonyID, -1, env.executorPrvKey)
		assert.Nil(t, err)
	}

	// We should now have 10 running processes
	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, numberOfProcesses)

	waitForProcesses(t, server, processes, core.FAILED)

	// We should now have 10 failed processes since max retries reached
	stat, err = client.ColonyStatistics(env.colonyID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, numberOfProcesses) // NOTE Failed!!

	server.Shutdown()
	<-done
}
