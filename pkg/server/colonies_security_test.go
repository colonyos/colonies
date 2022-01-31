package server

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestAddColonySecurity(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()

	privateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := crypto.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, "invalid_api_key")
	assert.NotNilf(t, err, "it should be possible to create a colony without correct api key")

	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteColonySecurity(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()

	privateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := crypto.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	err = client.DeleteColony(colony.ID, invalidPrivateKey)
	assert.NotNil(t, err)

	err = client.DeleteColony(colony.ID, serverPrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetColoniesSecurity(t *testing.T) {
	_, client, server, serverPrvKey, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now, try to get colonies info using an invalid api
	_, err := client.GetColonies(core.GenerateRandomID())
	assert.NotNil(t, err) // Should not work

	// Now, try to get colonies info using an invalid api
	_, err = client.GetColonies(serverPrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetColonyByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now, try to get colony1 info using colony2 credentials
	_, err := client.GetColonyByID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to get colony1 info using colony1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should work, cannot use colony1PrvKey as credential

	// Now, try to get colony1 info using runtime1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now, try to get colony1 info using runtime1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestAddRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	runtime3, _, _ := generateRuntime(t, env.colony1ID)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is bound to colony1, but not yet a member

	// Now, try to add runtime 3 to colony1 using colony 2 credentials
	_, err := client.AddRuntime(runtime3, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to add runtime 3 to colony1 using colony 1 credentials
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetRuntimesByColonySecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now try to access runtime1 using credential of runtime2
	_, err := client.GetRuntimes(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access runtime1 using runtime1 credential
	_, err = client.GetRuntimes(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntimes(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work, colony owner can also get runtimes

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntimes(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work, cannot use colony2 credential

	server.Shutdown()
	<-done
}

func TestGetRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now try to access runtime1 using credentials of runtime2
	_, err := client.GetRuntime(env.runtime1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access runtime1 using runtime1 credential
	_, err = client.GetRuntime(env.runtime1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntime(env.runtime1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should work, cannot use colony1 crendential

	// Now try to access runtime1 using colony1 credential
	_, err = client.GetRuntime(env.runtime1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestApproveRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	runtime3, _, _ := generateRuntime(t, env.colony1ID)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is a not yet approved member of colony 1

	_, err := client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime3.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestRejectRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)
	runtime3, _, _ := generateRuntime(t, env.colony1ID)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is a not yet approved member of colony 1

	_, err := client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.RejectRuntime(runtime3.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RejectRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestNonApprovedRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Add another runtime to colony1 and list all runtimes, it should be possible
	runtime3, _, runtime3PrvKey := generateRuntime(t, env.colony1ID)
	_, err := client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetRuntimes(env.colony1ID, runtime3PrvKey)
	assert.Nil(t, err) // Should work, runtime should be able to list all runtimes even if it is not approved

	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetRuntimes(env.colony1ID, runtime3PrvKey)
	assert.Nil(t, err) // Should also work

	server.Shutdown()
	<-done
}

func TestDeleteRuntimeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Add another runtime to colony1 and list all runtimes, it should be possible
	runtime3, _, runtime3PrvKey := generateRuntime(t, env.colony1ID)
	_, err := client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteRuntime(runtime3.ID, runtime3PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestSubmitProcessSpecSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec1 := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	_, err := client.SubmitProcessSpec(processSpec1, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.SubmitProcessSpec(processSpec1, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work, runtiume2 is not member of colony1

	processSpec2 := core.CreateProcessSpec(env.colony2ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

	processSpec1 := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	_, err := client.SubmitProcessSpec(processSpec1, env.runtime1PrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(env.colony2ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

func TestGetWaitingProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

func TestRunningProcessesSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	numberOfRunningProcesses := 2
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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
		processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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
		processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

	processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

	processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

	processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseSuccessful(processFromServer.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.CloseSuccessful(processFromServer.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Add another runtime to colony1 and try to close the process statred by runtime1, it should not be possible
	runtime3, _, runtime3PrvKey := generateRuntime(t, env.colony1ID)
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

	processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	_, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	processFromServer, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	err = client.CloseFailed(processFromServer.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.CloseFailed(processFromServer.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Add another runtime to colony1 and try to close the process started by runtime1, it should not be possible
	runtime3, _, runtime3PrvKey := generateRuntime(t, env.colony1ID)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.CloseFailed(processFromServer.ID, runtime3PrvKey)
	assert.NotNil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestAddAttributeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Add another runtime to colony1 and try to set an attribute in the assigned processes assigned to
	// runtime1, it should not be possible
	runtime3, _, runtime3PrvKey := generateRuntime(t, env.colony1ID)
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime3.ID, env.colony1PrvKey)
	assert.Nil(t, err)
	_, err = client.AddAttribute(attribute, runtime3PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddAttribute(attribute, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetAttributeSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtime1PrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.State)

	assignedProcess, err := client.AssignProcess(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", "helloworld")
	_, err = client.AddAttribute(attribute, env.runtime1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetAttribute(attribute.ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetAttribute(attribute.ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

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
