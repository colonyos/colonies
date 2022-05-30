package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubmitProcessSpecSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec1 := utils.CreateTestProcessSpec(env.colony1ID)
	_, err := client.SubmitProcessSpec(processSpec1, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.SubmitProcessSpec(processSpec1, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work, runtiume2 is not member of colony1

	processSpec2 := utils.CreateTestProcessSpec(env.colony2ID)
	_, err = client.SubmitProcessSpec(processSpec2, env.runtime2PrvKey)
	assert.Nil(t, err)

	_, err = client.SubmitProcessSpec(processSpec2, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work, runtiume1 is not member of colony2

	server.Shutdown()
	<-done
}

func TestAssignProcessSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec1 := utils.CreateTestProcessSpec(env.colony1ID)
	_, err := client.SubmitProcessSpec(processSpec1, env.runtime1PrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := utils.CreateTestProcessSpec(env.colony2ID)
	_, err = client.SubmitProcessSpec(processSpec2, env.runtime2PrvKey)
	assert.Nil(t, err)

	// Now try to assign a process from colony2 using runtime1 credentials
	_, err = client.AssignProcess(env.colony2ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to assign a process from colony2 using runtime1 credentials
	_, err = client.AssignProcess(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to assign a process from colony2 using runtime1 credentials
	_, err = client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.AssignProcess(env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work, only runtimes are allowed

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.AssignProcess(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work, only runtimes are allowed, also invalid credentials are used

	server.Shutdown()
	<-done
}

func TestGetProcessHistForColonySecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetProcessHistForColony(core.RUNNING, env.colony1ID, 60, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForColony(core.RUNNING, env.colony1ID, 60, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForColony(core.RUNNING, env.colony1ID, 60, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessHistForColony(core.RUNNING, env.colony1ID, 60, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetProcessHistForRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetProcessHistForRuntime(core.RUNNING, env.colony1ID, env.runtime1ID, 60, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForRuntime(core.RUNNING, env.colony1ID, env.runtime1ID, 60, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForRuntime(core.RUNNING, env.colony1ID, env.runtime1ID, 60, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessHistForRuntime(core.RUNNING, env.colony1ID, env.runtime1ID, 60, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetWaitingProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetRunningProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetRunningProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetRunningProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetWaitingProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetWaitingProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetSuccessfulProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
		assert.Nil(t, err)
		err = client.CloseSuccessful(processFromServer.ID, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetSuccessfulProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSuccessfulProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetFailedProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := utils.CreateTestProcessSpec(env.colony1ID)
		_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
		assert.Nil(t, err)
		err = client.CloseFailed(processFromServer.ID, env.runtime1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetFailedProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFailedProcesses(env.colony1ID, numberOfRunningProcesses, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetProcessSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcess(addedProcess.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcess(addedProcess.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteProcessSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteProcess(addedProcess.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteProcess(addedProcess.ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteProcess(addedProcess.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteProcess(addedProcess.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestDeleteAllProcessSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteAllProcesses(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteAllProcesses(env.colony1ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteAllProcesses(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteAllProcesses(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Shoul dwork

	server.Shutdown()
	<-done
}

func TestGetProcessStatSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	_, err := client.GetProcessStat(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessStat(env.colony2ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessStat(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessStat(env.colony2ID, env.runtime2PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessStat(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessStat(env.colony2ID, env.colony2PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessStat(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestCloseSuccessfulSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseSuccessful(processFromServer.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.CloseSuccessful(processFromServer.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Add another runtime to colony1 and try to close the process statred by runtime1, it should not be possible
	runtime3, runtime3PrvKey, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.CloseSuccessful(processFromServer.ID, runtime3PrvKey)
	assert.NotNil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestCloseFailedSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := utils.CreateTestProcessSpec(env.colony1ID)
	_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseFailed(processFromServer.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.CloseFailed(processFromServer.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Add another runtime to colony1 and try to close the process started by runtime1, it should not be possible
	runtime3, runtime3PrvKey, err := utils.CreateTestRuntimeWithKey(env.colony1ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.CloseFailed(processFromServer.ID, runtime3PrvKey)
	assert.NotNil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestSubscribeProcessesSecurity(t *testing.T) {
	_, client, server, _, done := setupTestEnv1(t)

	runtimeType := "test_runtime_type"

	crypto := crypto.CreateCrypto()
	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	subscription, err := client.SubscribeProcesses(runtimeType, core.WAITING, 100, invalidPrivateKey)
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
	assert.NotNil(t, err) // Should not work, we should have got an error "runtime not found"

	server.Shutdown()
	<-done
}

func TestSubscribeChangeStateProcessSecurity(t *testing.T) {
	_, client, server, _, done := setupTestEnv1(t)

	crypto := crypto.CreateCrypto()
	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	subscription, err := client.SubscribeProcess(core.GenerateRandomID(), core.WAITING, 100, invalidPrivateKey)
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
	assert.NotNil(t, err) // Should not work, we should have got an error "runtime not found"

	server.Shutdown()
	<-done
}
