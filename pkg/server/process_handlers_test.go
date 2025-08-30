package server

import (
	"context"
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
	funcSpec1 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, in)
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, funcSpec2.Conditions.ColonyName, addedProcess2.FunctionSpec.Conditions.ColonyName)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.colonyName, "", "", "", 100, env.executorPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArraysEqual(processes, processesFromServer))

	server.Shutdown()
	<-done
}

func TestSubmitProcessInvalidPriority(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	funcSpec := utils.CreateTestFunctionSpecWithEnv(env.colonyName, in)
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
	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessWithContext(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	_, err := client.AssignWithContext(env.colonyName, 100, ctx, "", "", env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestAssignProcessWithTimeout(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	addedProcessChan := make(chan *core.Process)
	go func() {
		time.Sleep(1 * time.Second)
		funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
		addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		addedProcessChan <- addedProcess
	}()

	// This function call will block for 60 seconds or until the Go-routine above submits a process spec
	assignProcess, err := client.Assign(env.colonyName, 60, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignProcess)

	addedProcess := <-addedProcessChan
	assert.Equal(t, addedProcess.ID, assignProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessWithTimeoutFail(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	_, err := client.Assign(env.colonyName, 1, "", "", env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestAssignProcessNoPriority(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec1.Priority = 0
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec2.Priority = 0
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec3.Priority = 0
	addedProcess3, err := client.Submit(funcSpec3, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessPriority(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec1.Priority = 1
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec2.Priority = 2
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec3.Priority = 5
	addedProcess3, err := client.Submit(funcSpec3, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec4.Priority = 5
	addedProcess4, err := client.Submit(funcSpec4, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessWithLimits(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec1.Conditions.CPU = "4000m"
	funcSpec1.Conditions.Memory = "4000G"

	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	addedProcess1, err = client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "1000m", "10G", env.executorPrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "1000m", "4000G", env.executorPrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "4000m", "3000G", env.executorPrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "4000m", "4000G", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessByNameAndLimits(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec1.Conditions.ExecutorNames = []string{"executor1"}
	funcSpec1.Conditions.CPU = "4000m"
	funcSpec1.Conditions.Memory = "4000G"
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec2.Conditions.CPU = "4000m"
	funcSpec2.Conditions.Memory = "4000G"
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec3.Conditions.CPU = "4000m"
	funcSpec3.Conditions.Memory = "4000G"
	addedProcess3, err := client.Submit(funcSpec3, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec4.Conditions.ExecutorNames = []string{"executor2"}
	funcSpec4.Conditions.CPU = "4000m"
	funcSpec4.Conditions.Memory = "4000G"
	addedProcess4, err := client.Submit(funcSpec4, env.executorPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor1.Name = "executor1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor1.Name, env.colonyPrvKey)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor2.Name = "executor2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor2.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "1000m", "1000G", executor2PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "1000m", "5000G", executor2PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "5000m", "1000G", executor2PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "5000m", "5000G", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "5000m", "5000Gi", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "4000m", "4000G", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	_, err = client.Assign(env.colonyName, -1, "4000m", "4000G", executor1PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "10000m", "9000G", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignProcessByName(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec1.Conditions.ExecutorNames = []string{"executor1"}
	addedProcess1, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	addedProcess2, err := client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	addedProcess3, err := client.Submit(funcSpec3, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.colonyName, make(map[string]string))
	funcSpec4.Conditions.ExecutorNames = []string{"executor2"}
	addedProcess4, err := client.Submit(funcSpec4, env.executorPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor1.Name = "executor1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor1.Name, env.colonyPrvKey)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor2.Name = "executor2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor2.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	_, err = client.Assign(env.colonyName, -1, "", "", executor1PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.colonyName, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestMarkAlive(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	executorFromServer, err := client.GetExecutor(env.colonyName, executor.Name, executorPrvKey)
	assert.Nil(t, err)

	time1 := executorFromServer.LastHeardFromTime
	time.Sleep(1 * time.Second)

	client.Assign(env.colonyName, -1, "", "", executorPrvKey) // This will update the last heard from

	executorFromServer, err = client.GetExecutor(env.colonyName, executor.Name, executorPrvKey)
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
		funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}

	// Get processes for the last 60 seconds
	processesFromServer, err := client.GetProcessHistForColony(core.WAITING, env.colonyName, 60, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	server.Shutdown()
	<-done
}

func TestGetProcessHistForExecutor(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	numberOfRunningProcesses := 10
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colony1Name)
		_, err := client.Submit(funcSpec, env.executor1PrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.colony1Name, -1, "", "", env.executor1PrvKey)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	funcSpec := utils.CreateTestFunctionSpec(env.colony1Name)
	_, err := client.Submit(funcSpec, env.executor1PrvKey)
	assert.Nil(t, err)
	_, err = client.Assign(env.colony1Name, -1, "", "", env.executor1PrvKey)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	// Get processes for the 60 seconds
	processesFromServer, err := client.GetProcessHistForExecutor(core.RUNNING, env.colony1Name, env.executor1ID, 60, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses+1)

	// Get processes for the last 2 seconds
	processesFromServer, err = client.GetProcessHistForExecutor(core.RUNNING, env.colony1Name, env.executor1ID, 2, env.executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 1)

	server.Shutdown()
	<-done
}

func TestGetWaitingProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetWaitingProcesses(env.colonyName, "", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetWaitingProcesses(env.colonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetWaitingProcesses(env.colonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetWaitingProcesses(env.colonyName, "", "", "", 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetRunningProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	executor1, executorPrvKey1, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor1.Type = "test_executor_type_1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor1.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	executor2, executorPrvKey2, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor2.Type = "test_executor_type_2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor2.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}
	for i := 0; i < numberOfRunningProcesses; i++ {
		if i < 5 {
			_, err = client.Assign(env.colonyName, -1, "", "", executorPrvKey1)
			assert.Nil(t, err)
		} else {
			_, err = client.Assign(env.colonyName, -1, "", "", executorPrvKey2)
			assert.Nil(t, err)
		}
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetRunningProcesses(env.colonyName, "", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetRunningProcesses(env.colonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetRunningProcesses(env.colonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetRunningProcesses(env.colonyName, "", "", "", 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetSuccessfulProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	executor1, executorPrvKey1, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor1.Type = "test_executor_type_1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor1.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	executor2, executorPrvKey2, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor2.Type = "test_executor_type_2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor2.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}
	var processFromServer *core.Process
	for i := 0; i < numberOfRunningProcesses; i++ {
		if i < 5 {
			processFromServer, err = client.Assign(env.colonyName, -1, "", "", executorPrvKey1)
			assert.Nil(t, err)
			err = client.Close(processFromServer.ID, executorPrvKey1)
			assert.Nil(t, err)
		} else {
			processFromServer, err = client.Assign(env.colonyName, -1, "", "", executorPrvKey2)
			assert.Nil(t, err)
			err = client.Close(processFromServer.ID, executorPrvKey2)
			assert.Nil(t, err)
		}
	}

	processesFromServer, err := client.GetSuccessfulProcesses(env.colonyName, "", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetSuccessfulProcesses(env.colonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetSuccessfulProcesses(env.colonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetSuccessfulProcesses(env.colonyName, "", "", "", 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetFailedProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	executor1, executorPrvKey1, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor1.Type = "test_executor_type_1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor1.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	executor2, executorPrvKey2, err := utils.CreateTestExecutorWithKey(env.colonyName)
	executor2.Type = "test_executor_type_2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor2.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
	}

	var processFromServer *core.Process
	for i := 0; i < numberOfRunningProcesses; i++ {
		if i < 5 {
			processFromServer, err = client.Assign(env.colonyName, -1, "", "", executorPrvKey1)
			assert.Nil(t, err)
			err = client.Fail(processFromServer.ID, []string{"error"}, executorPrvKey1)
			assert.Nil(t, err)
		} else {
			processFromServer, err = client.Assign(env.colonyName, -1, "", "", executorPrvKey2)
			assert.Nil(t, err)
			err = client.Fail(processFromServer.ID, []string{"error"}, executorPrvKey2)
			assert.Nil(t, err)
		}
	}

	processesFromServer, err := client.GetFailedProcesses(env.colonyName, "", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetFailedProcesses(env.colonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetFailedProcesses(env.colonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetFailedProcesses(env.colonyName, "", "", "", 10, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.executorPrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestRemoveProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.executorPrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	err = client.RemoveProcess(addedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err = client.GetProcess(addedProcess.ID, env.executorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, processFromServer)

	server.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColony(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colony1Name)
	addedProcess1, err := client.Submit(funcSpec1, env.executor1PrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colony2Name)
	addedProcess2, err := client.Submit(funcSpec2, env.executor2PrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess1.ID, env.executor1PrvKey)
	assert.True(t, addedProcess1.Equals(processFromServer))

	err = client.RemoveAllProcesses(env.colony1Name, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcess(addedProcess1.ID, env.executor1PrvKey)
	assert.NotNil(t, err)

	processFromServer, err = client.GetProcess(addedProcess2.ID, env.executor2PrvKey)
	assert.Nil(t, err)
	assert.True(t, addedProcess2.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateWaiting(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 2)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.RemoveAllProcessesWithState(env.colonyName, core.WAITING, env.colonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 0)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	server.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateRunning(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 1)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.RemoveAllProcessesWithState(env.colonyName, core.RUNNING, env.colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateSuccessful(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.Close(process.ID, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 1)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.RemoveAllProcessesWithState(env.colonyName, core.SUCCESS, env.colonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	server.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec1, env.executorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.colonyName)
	_, err = client.Submit(funcSpec2, env.executorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	err = client.Fail(process.ID, []string{"error"}, env.executorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 1)

	err = client.RemoveAllProcessesWithState(env.colonyName, core.FAILED, env.colonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
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

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
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

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	function := &core.Function{ColonyName: env.colonyName, ExecutorName: env.executorName, FuncName: funcSpec.FuncName}
	_, err = client.AddFunction(function, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Close(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.State)

	functions, err := client.GetFunctionsByExecutor(env.colonyName, env.executorName, env.executorPrvKey)
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

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
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

func TestSetOutput(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	_, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)

	output := make([]interface{}, 2)
	output[0] = "result1"
	err = client.SetOutput(assignedProcess.ID, output, env.executorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.executorPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, processFromServer.Output[0], "result1")

	server.Shutdown()
	<-done
}

func TestCloseFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
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

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec.MaxWaitTime = 1 // 1 second

	process, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	var processes []*core.Process
	processes = append(processes, process)
	waitForProcesses(t, server, processes, core.FAILED)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, 1)

	server.Shutdown()
	<-done
}

func TestMaxExecTime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec.MaxExecTime = 1 // 1 second

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	waitForProcesses(t, server, processes, core.WAITING)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeUnlimtedMaxRetries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec.MaxExecTime = 1 // 1 second
	funcSpec.MaxRetries = -1 // Unlimted number of retries

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	waitForProcesses(t, server, processes, core.WAITING)

	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	// Assign again
	for i := 0; i < numberOfProcesses; i++ {
		_, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
		assert.Nil(t, err)
	}

	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, numberOfProcesses)

	waitForProcesses(t, server, processes, core.WAITING)

	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeMaxRetries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	funcSpec.MaxExecTime = 3 // 3 seconds
	funcSpec.MaxRetries = 1  // Max 1 retries

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.executorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	waitForProcesses(t, server, processes, core.WAITING)

	// We should now have 10 waiting processes
	stat, err := client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	// Assign again
	for i := 0; i < numberOfProcesses; i++ {
		_, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
		assert.Nil(t, err)
	}

	// We should now have 10 running processes
	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, numberOfProcesses)

	waitForProcesses(t, server, processes, core.FAILED)

	// We should now have 10 failed processes since max retries reached
	stat, err = client.ColonyStatistics(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, numberOfProcesses) // NOTE Failed!!

	server.Shutdown()
	<-done
}

func TestPauseResumeAssignments(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Test pause assignments
	err := client.PauseColonyAssignments(env.colonyName, env.colonyPrvKey)
	assert.Nil(t, err)

	// Submit a process to test assignment blocking
	funcSpec := utils.CreateTestFunctionSpec(env.colonyName)
	addedProcess, err := client.Submit(funcSpec, env.executorPrvKey)
	assert.Nil(t, err)

	// Try to assign process - should fail because assignments are paused
	_, err = client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.NotNil(t, err)

	// Test resume assignments
	err = client.ResumeColonyAssignments(env.colonyName, env.colonyPrvKey)
	assert.Nil(t, err)

	// Now assignment should work
	assignedProcess, err := client.Assign(env.colonyName, -1, "", "", env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestPauseResumeAssignmentsColonyIsolation(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	// Create first colony
	colony1, colony1PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colony1PrvKey)
	assert.Nil(t, err)

	// Create second colony
	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	// Submit processes to both colonies
	funcSpec1 := utils.CreateTestFunctionSpec(colony1.Name)
	process1, err := client.Submit(funcSpec1, executor1PrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(colony2.Name)
	process2, err := client.Submit(funcSpec2, executor2PrvKey)
	assert.Nil(t, err)

	// Both colonies should allow assignments initially
	assignedProcess1, err := client.Assign(colony1.Name, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess1)

	assignedProcess2, err := client.Assign(colony2.Name, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess2)

	// Close the assigned processes to clean up
	err = client.Close(assignedProcess1.ID, executor1PrvKey)
	assert.Nil(t, err)
	err = client.Close(assignedProcess2.ID, executor2PrvKey)
	assert.Nil(t, err)

	// Submit new processes for the pause test
	process1, err = client.Submit(funcSpec1, executor1PrvKey)
	assert.Nil(t, err)
	process2, err = client.Submit(funcSpec2, executor2PrvKey)
	assert.Nil(t, err)

	// Pause assignments for colony1 only
	err = client.PauseColonyAssignments(colony1.Name, colony1PrvKey)
	assert.Nil(t, err)

	// Colony1 should be blocked from assignments
	_, err = client.Assign(colony1.Name, -1, "", "", executor1PrvKey)
	assert.NotNil(t, err)

	// Colony2 should still allow assignments (not affected by colony1's pause)
	assignedProcess2, err = client.Assign(colony2.Name, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess2)
	assert.Equal(t, process2.ID, assignedProcess2.ID)

	// Resume assignments for colony1
	err = client.ResumeColonyAssignments(colony1.Name, colony1PrvKey)
	assert.Nil(t, err)

	// Now colony1 should allow assignments again
	assignedProcess1, err = client.Assign(colony1.Name, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess1)
	assert.Equal(t, process1.ID, assignedProcess1.ID)

	// Colony2 should still work normally
	err = client.Close(assignedProcess2.ID, executor2PrvKey)
	assert.Nil(t, err)
	process2, err = client.Submit(funcSpec2, executor2PrvKey)
	assert.Nil(t, err)
	assignedProcess2, err = client.Assign(colony2.Name, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignedProcess2)

	// Clean up
	err = client.Close(assignedProcess1.ID, executor1PrvKey)
	assert.Nil(t, err)
	err = client.Close(assignedProcess2.ID, executor2PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
