package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	server.Shutdown()
	<-done
}

func TestDeleteColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 1)

	err = client.DeleteColony(addedColony.ID, serverPrvKey)
	assert.Nil(t, err)

	coloniesFromServer, err = client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.Len(t, coloniesFromServer, 0)

	server.Shutdown()
	<-done
}

func TestGetColony(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colony.ID, runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony1, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	var colonies []*core.Colony
	colonies = append(colonies, colony1)
	colonies = append(colonies, colony2)

	coloniesFromServer, err := client.GetColonies(serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromServer))

	server.Shutdown()
	<-done
}

func TestAddRuntime(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	assert.Nil(t, err)
	addedRuntime, err := client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))
	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the state will change after it has been approved
	addedRuntime.State = core.APPROVED

	runtimeFromServer, err := client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromServer)
	assert.True(t, addedRuntime.Equals(runtimeFromServer))

	server.Shutdown()
	<-done
}

func TestGetRuntimes(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime1, runtime1PrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	_, err = client.AddRuntime(runtime1, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime1.ID, colonyPrvKey)
	assert.Nil(t, err)

	runtime2, _, err := utils.CreateTestRuntimeWithKey(colony.ID)
	_, err = client.AddRuntime(runtime2, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime2.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the state will change after it has been approved
	runtime1.State = core.APPROVED
	runtime2.State = core.APPROVED

	var runtimes []*core.Runtime
	runtimes = append(runtimes, runtime1)
	runtimes = append(runtimes, runtime2)

	runtimesFromServer, err := client.GetRuntimes(colony.ID, runtime1PrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimes, runtimesFromServer))

	server.Shutdown()
	<-done
}

func TestApproveRejectRuntime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Add an approved runtime to use for the test below
	approvedRuntime, approvedRuntimePrvKey, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(approvedRuntime, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(approvedRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	testRuntime, _, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(testRuntime, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err := client.GetRuntime(testRuntime.ID, approvedRuntimePrvKey)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	err = client.ApproveRuntime(testRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntime(testRuntime.ID, approvedRuntimePrvKey)
	assert.Nil(t, err)
	assert.True(t, runtimeFromServer.IsApproved())

	err = client.RejectRuntime(testRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntime(testRuntime.ID, approvedRuntimePrvKey)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestDeleteRuntime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(env.colonyID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	// Try to get it
	runtimeFromServer, err := client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromServer)
	assert.True(t, runtime.ID == runtimeFromServer.ID)

	// Now delete it
	err = client.DeleteRuntime(runtime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	// Try to get it again, it should be gone
	runtimeFromServer, err = client.GetRuntime(runtime.ID, runtimePrvKey)
	assert.NotNil(t, err)
	assert.Nil(t, runtimeFromServer)

	server.Shutdown()
	<-done
}

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

	processSpec1 := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := utils.CreateTestProcessSpecWithEnv(env.colonyID, make(map[string]string))
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
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
		err = client.CloseFailed(processFromServer.ID, env.runtimePrvKey)
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

func TestGetProcessStat(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Waiting
	numberOfWaitingProcesses := 5
	for i := 0; i < numberOfWaitingProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Running
	numberOfRunningProcesses := 6
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
	}

	// Successful
	numberOfSuccessfulProcesses := 7
	for i := 0; i < numberOfSuccessfulProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseSuccessful(processFromServer.ID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Failed
	numberOfFailedProcesses := 8
	for i := 0; i < numberOfFailedProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colonyID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.CloseFailed(processFromServer.ID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	stat, err := client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	assert.Equal(t, stat.Waiting, numberOfWaitingProcesses)
	assert.Equal(t, stat.Running, numberOfRunningProcesses)
	assert.Equal(t, stat.Success, numberOfSuccessfulProcesses)
	assert.Equal(t, stat.Failed, numberOfFailedProcesses)

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

	err = client.CloseFailed(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcess(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.FAILED, assignedProcessFromServer.State)

	server.Shutdown()
	<-done
}

func TestAddGetAttributes(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, addedAttribute.ID)

	assignedProcessFromServer, err := client.GetProcess(assignedProcess.ID, env.runtimePrvKey)

	out := make(map[string]string)
	for _, attribute := range assignedProcessFromServer.Attributes {
		out[attribute.Key] = attribute.Value
	}

	assert.Equal(t, "helloworld", out["result"])

	attributeFromServer, err := client.GetAttribute(attribute.ID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, attributeFromServer.ID)

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

	stat, err := client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Waiting, 10)

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

	stat, err := client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Waiting, 10)

	// Assign again
	for i := 0; i < 10; i++ {
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	stat, err = client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Running, 10)

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err = client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Waiting, 10)

	server.Shutdown()
	<-done
}

func TestMaxExecTimeMaxretries(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := utils.CreateTestProcessSpec(env.colonyID)
	processSpec.MaxExecTime = 1 // 1 second
	processSpec.MaxRetries = 1  // Max 1 retries

	for i := 0; i < 10; i++ {
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err := client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Waiting, 10)

	// Assign again
	for i := 0; i < 10; i++ {
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	stat, err = client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Running, 10)

	// Wait for the process to time out
	time.Sleep(5 * time.Second)

	stat, err = client.GetProcessStat(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, stat.Failed, 10) // NOTE Failed!!

	server.Shutdown()
	<-done
}
