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
	processSpec1 := utils.CreateTestProcessSpecWithEnv(env.colonyID, in)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey)
	assert.Nil(t, err)

	processSpec2 := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, processSpec2.Conditions.ColonyID, addedProcess2.ProcessSpec.Conditions.ColonyID)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArraysEqual(processes, processesFromServer))

	server.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)
	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, assignedProcess)
	assert.NotNil(t, err)

	processSpec1 := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := utils.CreateTestProcessSpecWithEnv(env.colonyID, make(map[string]string))
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestAssignLatestProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec1 := utils.CreateTestProcessSpec(env.colonyID)
	_, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := utils.CreateTestProcessSpecWithEnv(env.colonyID, make(map[string]string))
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignLatestProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestMarkAlive(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err := client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.Nil(t, err)

	time1 := runtimeFromServer.LastHeardFromTime
	time.Sleep(1 * time.Second)

	client.AssignProcess(env.colonyID, runtimePrvKey) // This will update the last heard from

	runtimeFromServer, err = client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.Nil(t, err)
	time2 := runtimeFromServer.LastHeardFromTime

	assert.True(t, time1 != time2)

	server.Shutdown()
	<-done
}

func TestGetProcessHistForColony(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	// Get processes for the 60 seconds
	processesFromServer, err := client.GetProcessHistForColony(core.WAITING, env.colonyID, 60, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses+1)

	// Get processes for the last second
	processesFromServer, err = client.GetProcessHistForColony(core.WAITING, env.colonyID, 1, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 1)

	server.Shutdown()
	<-done
}

func TestGetProcessHistForRuntime(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	numberOfRunningProcesses := 10
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	_, err = client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	// Get processes for the 60 seconds
	processesFromServer, err := client.GetProcessHistForRuntime(core.RUNNING, env.colony1ID, env.runtime1ID, 60, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses+1)

	// Get processes for the last second
	processesFromServer, err = client.GetProcessHistForRuntime(core.RUNNING, env.colony1ID, env.runtime1ID, 1, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 1)

	server.Shutdown()
	<-done
}

func TestGetWaitingProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetWaitingProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetWaitingProcesses(env.colonyID, 10, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetRunningProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetRunningProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetRunningProcesses(env.colonyID, 10, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetSuccessfulProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseSuccessful(processFromServer.ID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetSuccessfulProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetSuccessfulProcesses(env.colonyID, 10, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetFailedProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseFailed(processFromServer.ID, "error", env.runtimePrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetFailedProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetFailedProcesses(env.colonyID, 10, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.runtimePrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestDeleteProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess.ID, env.runtimePrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	err = client.DeleteProcess(addedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	processFromServer, err = client.GetProcess(addedProcess.ID, env.runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, processFromServer)

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessesForColony(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	processSpec1 := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtime1PrvKey)
	assert.Nil(t, err)

	processSpec2 := utils.CreateTestProcessSpec(env.colony2ID)
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtime2PrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(addedProcess1.ID, env.runtime1PrvKey)
	assert.True(t, addedProcess1.Equals(processFromServer))

	err = client.DeleteAllProcesses(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcess(addedProcess1.ID, env.runtime1PrvKey)
	assert.NotNil(t, err)

	processFromServer, err = client.GetProcess(addedProcess2.ID, env.runtime2PrvKey)
	assert.Nil(t, err)
	assert.True(t, addedProcess2.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestCloseSuccessful(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.CloseSuccessful(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.State)

	server.Shutdown()
	<-done
}

func TestCloseFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.State)

	err = client.CloseFailed(assignedProcess.ID, "error", env.runtimePrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcess(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, processFromServer.State, core.FAILED)
	assert.Equal(t, processFromServer.ErrorMsg, "error")

	server.Shutdown()
	<-done
}

// Runtime 2 subscribes on process events and expects to receive an event when a new process is submitted
// Runtime 1 submitts a new process
// Runtime 2 receives an event
func TestSubscribeProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	runtimeType := "test_runtime_type"

	subscription, err := client.SubscribeProcesses(runtimeType, core.WAITING, 100, env.runtime2PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			waitForProcess <- err
		}
	}()

	time.Sleep(1 * time.Second)

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	_, err = client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

// Runtime 1 submits a process
// Runtime 2 subscribes on process events and expects to receive an event when the process finishes.
// Runtime 1 gets assign the process
// Runtime 1 finish the process
// Runtime 2 receives an event
func TestSubscribeChangeStateProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	subscription, err := client.SubscribeProcess(addedProcess.ID, core.SUCCESS, 100, env.runtime2PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			waitForProcess <- err
		}
	}()

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseSuccessful(assignedProcess.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

// Let change the order of the operations a bit, what about if the subscriber subscribes on an
// process state change event, but that event has already occurred. Then, the subscriber would what forever.
// The solution is to let the server send an event anyway if the wanted state is true already.
//
// Runtime 1 submits a process
// Runtime 1 gets assign the process
// Runtime 1 finish the process
// Runtime 2 subscribes on process events and expects to receive an event when the process finishes.
// Runtime 2 receives an event
func TestSubscribeChangeStateProcess2(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseSuccessful(assignedProcess.ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	time.Sleep(1 * time.Second)

	subscription, err := client.SubscribeProcess(addedProcess.ID, core.SUCCESS, 100, env.runtime2PrvKey)
	assert.Nil(t, err)

	waitForProcess := make(chan error)
	go func() {
		select {
		case <-subscription.ProcessChan:
			waitForProcess <- nil
		case err := <-subscription.ErrChan:
			waitForProcess <- err
		}
	}()

	err = <-waitForProcess
	assert.Nil(t, err)
	server.Shutdown()
	<-done
}

func TestMaxWaitTime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	processSpec.MaxWaitTime = 1 // 1 second

	_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, 1)

	server.Shutdown()
	<-done
}

func TestMaxExecTime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	processSpec.MaxExecTime = 1 // 1 second

	for i := 0; i < 10; i++ {
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 10)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeUnlimtedMaxretries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	processSpec.MaxExecTime = 1 // 1 second
	processSpec.MaxRetries = -1 // Unlimted number of retries

	for i := 0; i < 10; i++ {
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 10)

	// Assign again
	for i := 0; i < 10; i++ {
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	stat, err = client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, 10)

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err = client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 10)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeMaxretries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	processSpec.MaxExecTime = 2 // 1 second
	processSpec.MaxRetries = 1  // Max 1 retries

	for i := 0; i < 10; i++ {
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// We should now have 10 running processes

	// Wait for the process to time out
	time.Sleep(10 * time.Second)

	// We should now have 10 waiting processes
	stat, err := client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.WaitingProcesses, 10)

	// Assign again
	for i := 0; i < 10; i++ {
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	time.Sleep(1 * time.Second)

	// We should now have 10 running processes
	stat, err = client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.RunningProcesses, 10)

	// Wait for the process to time out
	time.Sleep(10 * time.Second)

	// We should now have 10 failed processes since max retries reached
	stat, err = client.ColonyStatistics(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.FailedProcesses, 10) // NOTE Failed!!

	server.Shutdown()
	<-done
}
