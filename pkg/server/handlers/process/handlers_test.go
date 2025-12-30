package process_test

import (
	"context"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)


func TestSubmitProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	funcSpec1 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, in)
	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, funcSpec2.Conditions.ColonyName, addedProcess2.FunctionSpec.Conditions.ColonyName)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArraysEqual(processes, processesFromServer))

	coloniesServer.Shutdown()
	<-done
}

func TestSubmitProcessInvalidPriority(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	funcSpec := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, in)
	funcSpec.Priority = constants.MIN_PRIORITY - 100
	_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	funcSpec.Priority = constants.MAX_PRIORITY + 100
	_, err = client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	addedProcess2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessWithContext(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	_, err := client.AssignWithContext(env.ColonyName, 100, ctx, "", "", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessWithTimeout(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	addedProcessChan := make(chan *core.Process)
	go func() {
		time.Sleep(1 * time.Second)
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		addedProcessChan <- addedProcess
	}()

	// This function call will block for 60 seconds or until the Go-routine above submits a process spec
	assignProcess, err := client.Assign(env.ColonyName, 60, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, assignProcess)

	addedProcess := <-addedProcessChan
	assert.Equal(t, addedProcess.ID, assignProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessWithTimeoutFail(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	_, err := client.Assign(env.ColonyName, 1, "", "", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessNoPriority(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.Priority = 0
	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec2.Priority = 0
	addedProcess2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec3.Priority = 0
	addedProcess3, err := client.Submit(funcSpec3, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessPriority(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.Priority = 1
	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec2.Priority = 2
	addedProcess2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec3.Priority = 5
	addedProcess3, err := client.Submit(funcSpec3, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec4.Priority = 5
	addedProcess4, err := client.Submit(funcSpec4, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessWithLimits(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.Conditions.CPU = "4000m"
	funcSpec1.Conditions.Memory = "4000G"

	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	addedProcess1, err = client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "1000m", "10G", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "1000m", "4000G", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "4000m", "3000G", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "4000m", "4000G", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessByNameAndLimits(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.Conditions.ExecutorNames = []string{"executor1"}
	funcSpec1.Conditions.CPU = "4000m"
	funcSpec1.Conditions.Memory = "4000G"
	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec2.Conditions.CPU = "4000m"
	funcSpec2.Conditions.Memory = "4000G"
	addedProcess2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec3.Conditions.CPU = "4000m"
	funcSpec3.Conditions.Memory = "4000G"
	addedProcess3, err := client.Submit(funcSpec3, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec4.Conditions.ExecutorNames = []string{"executor2"}
	funcSpec4.Conditions.CPU = "4000m"
	funcSpec4.Conditions.Memory = "4000G"
	addedProcess4, err := client.Submit(funcSpec4, env.ExecutorPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor1.Name = "executor1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor2.Name = "executor2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "1000m", "1000G", executor2PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "1000m", "5000G", executor2PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "5000m", "1000G", executor2PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "5000m", "5000G", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "5000m", "5000Gi", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "4000m", "4000G", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	_, err = client.Assign(env.ColonyName, -1, "4000m", "4000G", executor1PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "10000m", "9000G", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestAssignProcessByName(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.Conditions.ExecutorNames = []string{"executor1"}
	addedProcess1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	addedProcess2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec3 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	addedProcess3, err := client.Submit(funcSpec3, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec4 := utils.CreateTestFunctionSpecWithEnv(env.ColonyName, make(map[string]string))
	funcSpec4.Conditions.ExecutorNames = []string{"executor2"}
	addedProcess4, err := client.Submit(funcSpec4, env.ExecutorPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor1.Name = "executor1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor2.Name = "executor2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess3.ID, assignedProcess.ID)

	_, err = client.Assign(env.ColonyName, -1, "", "", executor1PrvKey)
	assert.NotNil(t, err)

	assignedProcess, err = client.Assign(env.ColonyName, -1, "", "", executor2PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess4.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestMarkAlive(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executorFromServer, err := client.GetExecutor(env.ColonyName, executor.Name, executorPrvKey)
	assert.Nil(t, err)

	time1 := executorFromServer.LastHeardFromTime
	time.Sleep(1 * time.Second)

	client.Assign(env.ColonyName, -1, "", "", executorPrvKey) // This will update the last heard from

	executorFromServer, err = client.GetExecutor(env.ColonyName, executor.Name, executorPrvKey)
	assert.Nil(t, err)
	time2 := executorFromServer.LastHeardFromTime

	assert.True(t, time1 != time2)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessHistForColony(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	numberOfRunningProcesses := 3
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// Get processes for the last 60 seconds
	processesFromServer, err := client.GetProcessHistForColony(core.WAITING, env.ColonyName, 60, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessHistForExecutor(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	numberOfRunningProcesses := 10
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	// Get processes for the 60 seconds
	processesFromServer, err := client.GetProcessHistForExecutor(core.RUNNING, env.Colony1Name, env.Executor1ID, 60, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses+1)

	// Get processes for the last 2 seconds
	processesFromServer, err = client.GetProcessHistForExecutor(core.RUNNING, env.Colony1Name, env.Executor1ID, 2, env.Executor1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 1)

	coloniesServer.Shutdown()
	<-done
}

func TestGetWaitingProcesses(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetWaitingProcesses(env.ColonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetWaitingProcesses(env.ColonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	coloniesServer.Shutdown()
	<-done
}

func TestGetRunningProcesses(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	executor1, executorPrvKey1, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor1.Type = "test_executor_type_1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executor2, executorPrvKey2, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor2.Type = "test_executor_type_2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}
	for i := 0; i < numberOfRunningProcesses; i++ {
		if i < 5 {
			_, err = client.Assign(env.ColonyName, -1, "", "", executorPrvKey1)
			assert.Nil(t, err)
		} else {
			_, err = client.Assign(env.ColonyName, -1, "", "", executorPrvKey2)
			assert.Nil(t, err)
		}
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetRunningProcesses(env.ColonyName, "", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetRunningProcesses(env.ColonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetRunningProcesses(env.ColonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetRunningProcesses(env.ColonyName, "", "", "", 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	coloniesServer.Shutdown()
	<-done
}

func TestGetSuccessfulProcesses(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	executor1, executorPrvKey1, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor1.Type = "test_executor_type_1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executor2, executorPrvKey2, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor2.Type = "test_executor_type_2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}
	var processFromServer *core.Process
	for i := 0; i < numberOfRunningProcesses; i++ {
		if i < 5 {
			processFromServer, err = client.Assign(env.ColonyName, -1, "", "", executorPrvKey1)
			assert.Nil(t, err)
			err = client.Close(processFromServer.ID, executorPrvKey1)
			assert.Nil(t, err)
		} else {
			processFromServer, err = client.Assign(env.ColonyName, -1, "", "", executorPrvKey2)
			assert.Nil(t, err)
			err = client.Close(processFromServer.ID, executorPrvKey2)
			assert.Nil(t, err)
		}
	}

	processesFromServer, err := client.GetSuccessfulProcesses(env.ColonyName, "", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetSuccessfulProcesses(env.ColonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetSuccessfulProcesses(env.ColonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetSuccessfulProcesses(env.ColonyName, "", "", "", 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	coloniesServer.Shutdown()
	<-done
}

func TestGetFailedProcesses(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	executor1, executorPrvKey1, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor1.Type = "test_executor_type_1"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor1.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executor2, executorPrvKey2, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	executor2.Type = "test_executor_type_2"
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
		if i < 5 {
			funcSpec.Conditions.ExecutorType = "test_executor_type_1"
		} else {
			funcSpec.Conditions.ExecutorType = "test_executor_type_2"
		}
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	var processFromServer *core.Process
	for i := 0; i < numberOfRunningProcesses; i++ {
		if i < 5 {
			processFromServer, err = client.Assign(env.ColonyName, -1, "", "", executorPrvKey1)
			assert.Nil(t, err)
			err = client.Fail(processFromServer.ID, []string{"error"}, executorPrvKey1)
			assert.Nil(t, err)
		} else {
			processFromServer, err = client.Assign(env.ColonyName, -1, "", "", executorPrvKey2)
			assert.Nil(t, err)
			err = client.Fail(processFromServer.ID, []string{"error"}, executorPrvKey2)
			assert.Nil(t, err)
		}
	}

	processesFromServer, err := client.GetFailedProcesses(env.ColonyName, "", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetFailedProcesses(env.ColonyName, "test_executor_type_1", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 5)

	processesFromServer, err = client.GetFailedProcesses(env.ColonyName, "test_executor_type_2", "", "", numberOfRunningProcesses, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 15)

	processesFromServer, err = client.GetFailedProcesses(env.ColonyName, "", "", "", 10, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.ExecutorPrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.ExecutorPrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	err = client.RemoveProcess(addedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processFromServer, err = client.GetProcess(addedProcess.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, processFromServer)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColony(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess1, err := client.Submit(funcSpec1, env.Executor1PrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.Colony2Name)
	addedProcess2, err := client.Submit(funcSpec2, env.Executor2PrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess1.ID, env.Executor1PrvKey)
	assert.True(t, addedProcess1.Equals(processFromServer))

	err = client.RemoveAllProcesses(env.Colony1Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcess(addedProcess1.ID, env.Executor1PrvKey)
	assert.NotNil(t, err)

	processFromServer, err = client.GetProcess(addedProcess2.ID, env.Executor2PrvKey)
	assert.Nil(t, err)
	assert.True(t, addedProcess2.Equals(processFromServer))

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateWaiting(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 2)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.RemoveAllProcessesWithState(env.ColonyName, core.WAITING, env.ColonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 0)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateRunning(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 1)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.RemoveAllProcessesWithState(env.ColonyName, core.RUNNING, env.ColonyPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateSuccessful(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 1)
	assert.Equal(t, stat.FailedProcesses, 0)

	err = client.RemoveAllProcessesWithState(env.ColonyName, core.SUCCESS, env.ColonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessesForColonyWithStateFailed(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	err = client.Fail(process.ID, []string{"error"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 1)

	err = client.RemoveAllProcessesWithState(env.ColonyName, core.FAILED, env.ColonyPrvKey)
	assert.Nil(t, err)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 1)
	assert.Equal(t, stat.RunningProcesses, 0)
	assert.Equal(t, stat.SuccessfulProcesses, 0)
	assert.Equal(t, stat.FailedProcesses, 0)

	coloniesServer.Shutdown()
	<-done
}

func TestCloseSuccessful(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.State)

	coloniesServer.Shutdown()
	<-done
}

func TestCloseSuccessfulWithFunctions(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	function := &core.Function{ColonyName: env.ColonyName, ExecutorName: env.ExecutorName, FuncName: funcSpec.FuncName}
	_, err = client.AddFunction(function, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	time.Sleep(100 * time.Millisecond)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.State)

	functions, err := client.GetFunctionsByExecutor(env.ColonyName, env.ExecutorName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, functions[0].Counter, 1)
	assert.Greater(t, functions[0].MinWaitTime, 0.0)
	assert.Greater(t, functions[0].MaxWaitTime, 0.0)
	assert.Greater(t, functions[0].MinExecTime, 0.0)
	assert.Greater(t, functions[0].MaxExecTime, 0.0)
	assert.Greater(t, functions[0].AvgWaitTime, 0.0)
	assert.Greater(t, functions[0].AvgExecTime, 0.0)

	coloniesServer.Shutdown()
	<-done
}

func TestCloseSuccessfulWithOutput(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	output := make([]interface{}, 2)
	output[0] = "result1"
	output[1] = "result2"
	err = client.CloseWithOutput(assignedProcess.ID, output, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assert.Len(t, processFromServer.Output, 2)
	assert.Equal(t, processFromServer.Output[0], "result1")
	assert.Equal(t, processFromServer.Output[1], "result2")

	coloniesServer.Shutdown()
	<-done
}

func TestSetOutput(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	output := make([]interface{}, 2)
	output[0] = "result1"
	err = client.SetOutput(assignedProcess.ID, output, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assert.Equal(t, processFromServer.Output[0], "result1")

	coloniesServer.Shutdown()
	<-done
}

func TestCloseFailed(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.Fail(assignedProcess.ID, []string{"error"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Equal(t, processFromServer.State, core.FAILED)
	assert.Len(t, processFromServer.Errors, 1)
	assert.Equal(t, processFromServer.Errors[0], "error")

	coloniesServer.Shutdown()
	<-done
}

func TestMaxWaitTime(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.MaxWaitTime = 1 // 1 second

	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	var processes []*core.Process
	processes = append(processes, process)
	server.WaitForProcesses(t, coloniesServer, processes, core.FAILED)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, 1)

	coloniesServer.Shutdown()
	<-done
}

func TestMaxExecTime(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.MaxExecTime = 1 // 1 second

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	server.WaitForProcesses(t, coloniesServer, processes, core.WAITING)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	coloniesServer.Shutdown()
	<-done
}

func TestMaxExecTimeUnlimtedMaxRetries(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.MaxExecTime = 1 // 1 second
	funcSpec.MaxRetries = -1 // Unlimted number of retries

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	server.WaitForProcesses(t, coloniesServer, processes, core.WAITING)

	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	// Assign again
	for i := 0; i < numberOfProcesses; i++ {
		_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, numberOfProcesses)

	server.WaitForProcesses(t, coloniesServer, processes, core.WAITING)

	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	coloniesServer.Shutdown()
	<-done
}

func TestMaxExecTimeMaxRetries(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec.MaxExecTime = 3 // 3 seconds
	funcSpec.MaxRetries = 1  // Max 1 retries

	numberOfProcesses := 10
	var processes []*core.Process
	for i := 0; i < numberOfProcesses; i++ {
		_, err := client.Submit(funcSpec, env.ExecutorPrvKey)
		assert.Nil(t, err)
		process, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
		processes = append(processes, process)
	}

	server.WaitForProcesses(t, coloniesServer, processes, core.WAITING)

	// We should now have 10 waiting processes
	stat, err := client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, numberOfProcesses)

	// Assign again
	for i := 0; i < numberOfProcesses; i++ {
		_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
		assert.Nil(t, err)
	}

	// We should now have 10 running processes
	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, numberOfProcesses)

	server.WaitForProcesses(t, coloniesServer, processes, core.FAILED)

	// We should now have 10 failed processes since max retries reached
	stat, err = client.ColonyStatistics(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, numberOfProcesses) // NOTE Failed!!

	coloniesServer.Shutdown()
	<-done
}

func TestPauseResumeAssignments(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Test pause assignments
	err := client.PauseColonyAssignments(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Submit a process to test assignment blocking
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to assign process - should fail because assignments are paused
	_, err = client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	// Test resume assignments
	err = client.ResumeColonyAssignments(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Now assignment should work
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess.ID, assignedProcess.ID)

	coloniesServer.Shutdown()
	<-done
}

func TestPauseResumeAssignmentsColonyIsolation(t *testing.T) {
	client, coloniesServer, serverPrvKey, done := server.PrepareTests(t)

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

	coloniesServer.Shutdown()
	<-done
}

// TestCloseSuccessfulOnFailedProcess tests that trying to close a failed process
// as successful returns a proper single error response (not concatenated JSON).
// This test verifies the fix for the double HandleHTTPError bug.
func TestCloseSuccessfulOnFailedProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit and assign a process
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess.ID, assignedProcess.ID)

	// Fail the process first
	err = client.Fail(assignedProcess.ID, []string{"intentional failure"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify process is in FAILED state
	processFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.FAILED, processFromServer.State)

	// Try to close the failed process as successful - this should fail with a proper error
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Closing a failed process as successful should return an error")
	assert.Contains(t, err.Error(), "Tried to set failed process as successful")

	coloniesServer.Shutdown()
	<-done
}

// TestSetOutputOnFailedProcess tests that trying to set output on a failed process
// returns a proper single error response (not concatenated JSON).
func TestSetOutputOnFailedProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit and assign a process
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	addedProcess, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess.ID, assignedProcess.ID)

	// Fail the process first
	err = client.Fail(assignedProcess.ID, []string{"intentional failure"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Verify process is in FAILED state
	processFromServer, err := client.GetProcess(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.FAILED, processFromServer.State)

	// Try to set output on the failed process - this should fail with a proper error
	output := make([]interface{}, 1)
	output[0] = "test output"
	err = client.SetOutput(assignedProcess.ID, output, env.ExecutorPrvKey)
	assert.NotNil(t, err, "Setting output on a failed process should return an error")

	coloniesServer.Shutdown()
	<-done
}

func TestGetPauseStatus(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Initially, assignments should not be paused
	isPaused, err := client.AreColonyAssignmentsPaused(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.False(t, isPaused)

	// Pause assignments
	err = client.PauseColonyAssignments(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Now the status should be paused
	isPaused, err = client.AreColonyAssignmentsPaused(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.True(t, isPaused)

	// Resume assignments
	err = client.ResumeColonyAssignments(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Status should be unpaused again
	isPaused, err = client.AreColonyAssignmentsPaused(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.False(t, isPaused)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessesByState(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit processes
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	process1, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	process2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Test getting waiting processes
	waitingProcesses, err := client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(waitingProcesses))

	// Assign first process
	assignedProcess, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process1.ID, assignedProcess.ID)

	// Test getting running processes
	runningProcesses, err := client.GetRunningProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(runningProcesses))

	// Test getting waiting processes (should be 1 now)
	waitingProcesses, err = client.GetWaitingProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(waitingProcesses))
	assert.Equal(t, process2.ID, waitingProcesses[0].ID)

	// Close successful
	err = client.Close(assignedProcess.ID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Test getting successful processes
	successfulProcesses, err := client.GetSuccessfulProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(successfulProcesses))
	assert.Equal(t, assignedProcess.ID, successfulProcesses[0].ID)

	// Assign and fail second process
	assignedProcess2, err := client.Assign(env.ColonyName, -1, "", "", env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process2.ID, assignedProcess2.ID)

	err = client.Fail(assignedProcess2.ID, []string{"test failure"}, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Test getting failed processes
	failedProcesses, err := client.GetFailedProcesses(env.ColonyName, "", "", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(failedProcesses))
	assert.Equal(t, assignedProcess2.ID, failedProcesses[0].ID)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessesWithLabel(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit processes with different labels
	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.Label = "label1"
	process1, err := client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec2.Label = "label2"
	process2, err := client.Submit(funcSpec2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Test getting processes with label filter
	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "label1", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(processes))
	assert.Equal(t, process1.ID, processes[0].ID)

	processes, err = client.GetWaitingProcesses(env.ColonyName, "", "label2", "", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(processes))
	assert.Equal(t, process2.ID, processes[0].ID)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessesWithInitiator(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Create a user in the colony
	user, userPrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Submit a process as the executor
	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	_, err = client.Submit(funcSpec1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Submit a process as the user
	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	process2, err := client.Submit(funcSpec2, userPrvKey)
	assert.Nil(t, err)

	// Get processes with initiator filter
	processes, err := client.GetWaitingProcesses(env.ColonyName, "", "", "test_user", 100, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(processes))
	assert.Equal(t, process2.ID, processes[0].ID)

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessesWithInvalidInitiator(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Try to get processes with a non-existent initiator
	_, err := client.GetWaitingProcesses(env.ColonyName, "", "", "nonexistent_user", 100, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestCloseFailedNotAssigned(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit a process
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to close the process as failed without assigning it first
	err = client.Fail(process.ID, []string{"test error"}, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestCloseSuccessfulNotAssigned(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit a process
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to close the process as successful without assigning it first
	err = client.Close(process.ID, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveProcessInWorkflow(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Create a simple workflow with 2 processes
	funcSpec1 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec1.NodeName = "task1"
	funcSpec2 := utils.CreateTestFunctionSpec(env.ColonyName)
	funcSpec2.NodeName = "task2"
	funcSpec2.Conditions.Dependencies = []string{"task1"}

	workflowSpec := core.CreateWorkflowSpec(env.ColonyName)
	workflowSpec.FunctionSpecs = append(workflowSpec.FunctionSpecs, *funcSpec1)
	workflowSpec.FunctionSpecs = append(workflowSpec.FunctionSpecs, *funcSpec2)

	graph, err := client.SubmitWorkflowSpec(workflowSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, graph)
	assert.Equal(t, 2, len(graph.ProcessIDs))

	// Try to remove a process that is part of the workflow - should fail
	err = client.RemoveProcess(graph.ProcessIDs[0], env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestSetOutputOnWaitingProcess(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Submit a process
	funcSpec := utils.CreateTestFunctionSpec(env.ColonyName)
	process, err := client.Submit(funcSpec, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to set output on a waiting process - should fail
	output := make([]interface{}, 1)
	output[0] = "test output"
	err = client.SetOutput(process.ID, output, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}

func TestCloseWithWrongExecutor(t *testing.T) {
	client, coloniesServer, serverPrvKey, done := server.PrepareTests(t)

	// Create colony
	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Create two executors
	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor1.Name, colonyPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor2.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Submit and assign a process to executor1
	funcSpec := utils.CreateTestFunctionSpec(colony.Name)
	process, err := client.Submit(funcSpec, executor1PrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.Assign(colony.Name, -1, "", "", executor1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process.ID, assignedProcess.ID)

	// Try to close the process with executor2 - should fail
	err = client.Close(assignedProcess.ID, executor2PrvKey)
	assert.NotNil(t, err)

	// Try to fail the process with executor2 - should fail
	err = client.Fail(assignedProcess.ID, []string{"error"}, executor2PrvKey)
	assert.NotNil(t, err)

	coloniesServer.Shutdown()
	<-done
}
