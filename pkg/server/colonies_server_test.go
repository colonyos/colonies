package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)

	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	colonyAdded, err := client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyAdded))

	server.Shutdown()
	<-done
}

func TestGetColony(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)

	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	_, err = client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	runtime, _, runtimePrvKey := generateRuntime(t, colonyID)
	_, err = client.AddRuntime(runtime, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colonyID, runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromServer))

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)

	prvKey1, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID1, err := security.GenerateID(prvKey1)
	assert.Nil(t, err)
	colony1 := core.CreateColony(colonyID1, "test_colony_name")
	_, err = client.AddColony(colony1, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	prvKey2, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID2, err := security.GenerateID(prvKey2)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colonyID2, "test_colony_name")
	_, err = client.AddColony(colony2, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	var colonies []*core.Colony
	colonies = append(colonies, colony1)
	colonies = append(colonies, colony2)

	coloniesFromServer, err := client.GetColonies(rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromServer))

	server.Shutdown()
	<-done
}

func TestAddRuntime(t *testing.T) {
	rootPassword := "testapikey"
	server, done := PrepareTests(t, rootPassword)

	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a runtime
	runtimePrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	runtimeID, err := security.GenerateID(runtimePrvKey)
	assert.Nil(t, err)

	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime := core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	addedRuntime, err := client.AddRuntime(runtime, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))

	runtimeFromServer, err := client.GetRuntime(runtimeID, runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromServer)
	assert.True(t, runtime.Equals(runtimeFromServer))

	server.Shutdown()
	<-done
}

func TestGetRuntimes(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)

	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a Runtime
	runtime1PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	runtime1ID, err := security.GenerateID(runtime1PrvKey)
	assert.Nil(t, err)

	name := "test_runtime_name_1"
	runtimeType := "test_runtime_type"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime1 := core.CreateRuntime(runtime1ID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime1, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a Runtime
	runtime2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	runtime2ID, err := security.GenerateID(runtime2PrvKey)
	assert.Nil(t, err)

	name = "test_runtime_name_2"
	runtime2 := core.CreateRuntime(runtime2ID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime2, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	var runtimes []*core.Runtime
	runtimes = append(runtimes, runtime1)
	runtimes = append(runtimes, runtime2)

	runtimesFromServer, err := client.GetRuntimes(colonyID, runtime1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimes, runtimesFromServer))

	server.Shutdown()
	<-done
}

func TestApproveRejectRuntime(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	runtimeFromServer, err := client.GetRuntime(env.runtimeID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	err = client.ApproveRuntime(env.runtime.ID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntime(env.runtimeID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, runtimeFromServer.IsApproved())

	err = client.RejectRuntime(env.runtime.ID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntime(env.runtimeID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestSubmitProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, in)
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, processSpec2.Conditions.ColonyID, addedProcess2.ProcessSpec.Conditions.ColonyID)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.colonyID, 100, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArraysEqual(processes, processesFromServer))

	server.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess1, err := client.SubmitProcessSpec(processSpec1, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess2, err := client.SubmitProcessSpec(processSpec2, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestGetWaitingProcesses(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetWaitingProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetWaitingProcesses(env.colonyID, 10, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetRunningProcesses(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
		_, err = client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetRunningProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetRunningProcesses(env.colonyID, 10, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetSuccessfulProcesses(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
		err = client.MarkSuccessful(processFromServer, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetSuccessfulProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetSuccessfulProcesses(env.colonyID, 10, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

func TestGetFailedProcesses(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	numberOfRunningProcesses := 20
	for i := 0; i < numberOfRunningProcesses; i++ {
		processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
		_, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
		processFromServer, err := client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
		assert.Nil(t, err)
		err = client.MarkFailed(processFromServer, env.runtimePrvKey)
		assert.Nil(t, err)
	}

	processesFromServer, err := client.GetFailedProcesses(env.colonyID, numberOfRunningProcesses, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, numberOfRunningProcesses)

	processesFromServer, err = client.GetFailedProcesses(env.colonyID, 10, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Len(t, processesFromServer, 10)

	server.Shutdown()
	<-done
}

////////////////////////////////////

func TestMarkSuccessful(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.colonyID, env.runtimePrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status)

	err = client.MarkSuccessful(assignedProcess, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID, env.colonyID, env.runtimePrvKey)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.Status)

	server.Shutdown()
	<-done
}

func TestMarkFailed(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.colonyID, env.runtimePrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status)

	err = client.MarkFailed(assignedProcess, env.runtimePrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID, env.colonyID, env.runtimePrvKey)
	assert.Equal(t, core.FAILED, assignedProcessFromServer.Status)

	server.Shutdown()
	<-done
}

func TestAddAttributes(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.SubmitProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.colonyID, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, addedAttribute.ID)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.colonyID, env.runtimePrvKey)

	out := make(map[string]string)
	for _, attribute := range assignedProcessFromServer.Attributes {
		out[attribute.Key] = attribute.Value
	}

	assert.Equal(t, "helloworld", out["result"])

	attributeFromServer, err := client.GetAttribute(attribute.ID, addedProcess.ID, env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, attributeFromServer.ID)

	server.Shutdown()
	<-done
}
