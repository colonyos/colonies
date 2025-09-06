package process_test

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestSubmitSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec1 := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec1, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.Submit(funcSpec1, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work, runtiume2 is not member of colony1

	funcSpec2 := utils.CreateTestFunctionSpec(env.Colony2Name)
	_, err = client.Submit(funcSpec2, env.Executor2PrvKey)
	assert.Nil(t, err)

	_, err = client.Submit(funcSpec2, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work, runtiume1 is not member of colony2

	coloniesServer.Shutdown()
	<-done
}

func TestAssignSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec1 := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec1, env.Executor1PrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	funcSpec2 := utils.CreateTestFunctionSpec(env.Colony2Name)
	_, err = client.Submit(funcSpec2, env.Executor2PrvKey)
	assert.Nil(t, err)

	// Now try to assign a process from colony2 using executor1 credentials
	_, err = client.Assign(env.Colony2Name, -1, "", "", env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to assign a process from colony2 using executor1 credentials
	_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to assign a process from colony2 using executor1 credentials
	_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.Assign(env.Colony1Name, -1, "", "", env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work, only executors are allowed

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.Assign(env.Colony1Name, -1, "", "", env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work, only executors are allowed, also invalid credentials are used

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessHistForColonySecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetProcessHistForColony(core.RUNNING, env.Colony1Name, 60, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForColony(core.RUNNING, env.Colony1Name, 60, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForColony(core.RUNNING, env.Colony1Name, 60, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessHistForColony(core.RUNNING, env.Colony1Name, 60, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessHistForExecutorSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetProcessHistForExecutor(core.RUNNING, env.Colony1Name, env.Executor1ID, 60, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForExecutor(core.RUNNING, env.Colony1Name, env.Executor1ID, 60, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcessHistForExecutor(core.RUNNING, env.Colony1Name, env.Executor1ID, 60, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.GetProcessHistForExecutor(core.RUNNING, env.Colony1Name, env.Executor1ID, 60, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	coloniesServer.Shutdown()
	<-done
}

func TestGetWaitingProcessesSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
		_, err = client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetRunningProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetRunningProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestGetRunningProcessesSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetWaitingProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetWaitingProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestGetSuccessfulProcessesSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
		assert.Nil(t, err)
		err = client.Close(processFromServer.ID, env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetSuccessfulProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetSuccessfulProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestGetFailedProcessesSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
		_, err := client.Submit(funcSpec, env.Executor1PrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
		assert.Nil(t, err)
		err = client.Fail(processFromServer.ID, []string{"error"}, env.Executor1PrvKey)
		assert.Nil(t, err)
	}

	_, err := client.GetFailedProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFailedProcesses(env.Colony1Name, "", "", "", numberOfRunningProcesses, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestGetProcessSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetProcess(addedProcess.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetProcess(addedProcess.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveProcessSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	addedProcess, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveProcess(addedProcess.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveProcess(addedProcess.ID, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveProcess(addedProcess.ID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveProcess(addedProcess.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestRemoveAllProcessSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.RemoveAllProcesses(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllProcesses(env.Colony1Name, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllProcesses(env.Colony1Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveAllProcesses(env.Colony1Name, env.Colony1PrvKey)
	assert.Nil(t, err) // Shoul dwork

	coloniesServer.Shutdown()
	<-done
}

func TestCloseSuccessfulSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.Close(processFromServer.ID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.Close(processFromServer.ID, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Add another executor to colony1 and try to close the process statred by executor1, it should not be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.Close(processFromServer.ID, executor3PrvKey)
	assert.NotNil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestSetOutputSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	output := make([]interface{}, 2)
	output[0] = "result1"

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.SetOutput(processFromServer.ID, output, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.SetOutput(processFromServer.ID, output, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.SetOutput(processFromServer.ID, output, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.SetOutput(processFromServer.ID, output, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestCloseFailedSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	funcSpec := utils.CreateTestFunctionSpec(env.Colony1Name)
	_, err := client.Submit(funcSpec, env.Executor1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.Assign(env.Colony1Name, -1, "", "", env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.Fail(processFromServer.ID, []string{"error"}, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.Fail(processFromServer.ID, []string{"error"}, env.Executor1PrvKey)
	assert.Nil(t, err) // Should work

	// Add another executor to colony1 and try to close the process started by executor1, it should not be possible
	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)
	err = client.Fail(processFromServer.ID, []string{"error"}, executor3PrvKey)
	assert.NotNil(t, err) // Should work

	coloniesServer.Shutdown()
	<-done
}

func TestPauseResumeAssignmentsSecurity(t *testing.T) {
	env, client, coloniesServer, _, done := server.SetupTestEnv2(t)

	// Test pause assignments with wrong key (should fail)
	err := client.PauseColonyAssignments(env.ColonyName, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	// Test resume assignments with wrong key (should fail)
	err = client.ResumeColonyAssignments(env.ColonyName, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	// Test with correct colony key (should work)
	err = client.PauseColonyAssignments(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	err = client.ResumeColonyAssignments(env.ColonyName, env.ColonyPrvKey)
	assert.Nil(t, err)

	coloniesServer.Shutdown()
	<-done
}
