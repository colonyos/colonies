package server

import (
	"colonies/pkg/core"
	"colonies/pkg/security/crypto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	client, server, serverKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()
	colonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := crypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	colonyAdded, err := client.AddColony(colony, serverKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyAdded))

	server.Shutdown()
	<-done
}

func TestGetColony(t *testing.T) {
	client, server, serverKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()
	colonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := crypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	_, err = client.AddColony(colony, serverKey)
	assert.Nil(t, err)

	runtime, _, runtimePrvKey := generateRuntime(t, colonyID)
	_, err = client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colonyID, runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()
	prvKey1, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID1, err := crypto.GenerateID(prvKey1)
	assert.Nil(t, err)
	colony1 := core.CreateColony(colonyID1, "test_colony_name")
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	prvKey2, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID2, err := crypto.GenerateID(prvKey2)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colonyID2, "test_colony_name")
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

	// Create a Colony
	crypto := crypto.CreateCrypto()
	colonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := crypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Create a runtime
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	assert.Nil(t, err)

	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime := core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	addedRuntime, err := client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))
	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the status will change after it has been approved
	addedRuntime.Status = core.APPROVED

	runtimeFromServer, err := client.GetRuntime(runtimeID, runtimePrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromServer)
	assert.True(t, addedRuntime.Equals(runtimeFromServer))

	server.Shutdown()
	<-done
}

func TestGetRuntimes(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	// Create a Colony
	crypto := crypto.CreateCrypto()
	colonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := crypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Create a Runtime
	runtime1PrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtime1ID, err := crypto.GenerateID(runtime1PrvKey)
	assert.Nil(t, err)

	name := "test_runtime_name_1"
	runtimeType := "test_runtime_type"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime1 := core.CreateRuntime(runtime1ID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime1, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime1.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Create a Runtime
	runtime2PrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtime2ID, err := crypto.GenerateID(runtime2PrvKey)
	assert.Nil(t, err)

	name = "test_runtime_name_2"
	runtime2 := core.CreateRuntime(runtime2ID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime2, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(runtime2.ID, colonyPrvKey)
	assert.Nil(t, err)

	// Just to make the comparison below work, the status will change after it has been approved
	runtime1.Status = core.APPROVED
	runtime2.Status = core.APPROVED

	var runtimes []*core.Runtime
	runtimes = append(runtimes, runtime1)
	runtimes = append(runtimes, runtime2)

	runtimesFromServer, err := client.GetRuntimes(colonyID, runtime1PrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimes, runtimesFromServer))

	server.Shutdown()
	<-done
}

func TestApproveRejectRuntime(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	// Add an approved runtime to use for the test below
	approvedRuntime, _, approvedRuntimePrvKey := generateRuntime(t, env.colonyID)
	_, err := client.AddRuntime(approvedRuntime, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveRuntime(approvedRuntime.ID, env.colonyPrvKey)
	assert.Nil(t, err)

	testRuntime, _, _ := generateRuntime(t, env.colonyID)
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

func TestSubmitProcess(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, in)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey)
	assert.Nil(t, err)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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

func TestGetWaitingProcesses(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
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
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.MarkSuccessful(processFromServer.ID, env.runtimePrvKey)
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
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
		assert.Nil(t, err)
		err = client.MarkFailed(processFromServer.ID, env.runtimePrvKey)
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

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcessByID(addedProcess.ID, env.runtimePrvKey)
	assert.True(t, addedProcess.Equals(processFromServer))

	server.Shutdown()
	<-done
}

func TestMarkSuccessful(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status)

	err = client.MarkSuccessful(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.Status)

	server.Shutdown()
	<-done
}

func TestMarkFailed(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status)

	err = client.MarkFailed(assignedProcess.ID, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID, env.runtimePrvKey)
	assert.Equal(t, core.FAILED, assignedProcessFromServer.Status)

	server.Shutdown()
	<-done
}

func TestAddGetAttributes(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, addedAttribute.ID)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.runtimePrvKey)

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
