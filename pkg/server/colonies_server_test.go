package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
	"colonies/pkg/logging"
	"colonies/pkg/security"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func PrepareTests(t *testing.T, rootPassword string) (*ColoniesServer, chan bool) {
	debug := false

	if debug {
		logging.DisableDebug()
	}

	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	server := CreateColoniesServer(db, 8080, rootPassword, "../../cert/key.pem", "../../cert/cert.pem", debug)
	done := make(chan bool)

	go func() {
		server.ServeForever()
		done <- true
	}()

	return server, done
}

func TestAddColony(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)

	prvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(prvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	colonyAdded, err := client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyAdded))

	colonyFromServer, err := client.GetColonyByID(colonyID, prvKey)
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

	name := "test_runtime"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime := core.CreateRuntime(runtimeID, name, colonyID, cpu, cores, mem, gpu, gpus)
	addedRuntime, err := client.AddRuntime(runtime, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, runtime.Equals(addedRuntime))

	runtimeFromServer, err := client.GetRuntimeByID(runtimeID, colonyID, colonyPrvKey, TESTHOST, TESTPORT)
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

	name := "test_runtime 1"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime1 := core.CreateRuntime(runtime1ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime1, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a Runtime
	runtime2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	runtime2ID, err := security.GenerateID(runtime2PrvKey)
	assert.Nil(t, err)

	name = "test_runtime 2"
	runtime2 := core.CreateRuntime(runtime2ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime2, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	var runtimes []*core.Runtime
	runtimes = append(runtimes, runtime1)
	runtimes = append(runtimes, runtime2)

	runtimesFromServer, err := client.GetRuntimesByColonyID(colonyID, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, core.IsRuntimeArraysEqual(runtimes, runtimesFromServer))

	server.Shutdown()
	<-done
}

type clientTestEnv struct {
	colonyID      string
	colony        *core.Colony
	colonyPrvKey  string
	runtimeID     string
	runtime       *core.Runtime
	runtimePrvKey string
}

func createTestEnv(t *testing.T, rootPassword string) *clientTestEnv {
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

	name := "test_runtime"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime := core.CreateRuntime(runtimeID, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	return &clientTestEnv{colonyID: colonyID,
		colony:        colony,
		colonyPrvKey:  colonyPrvKey,
		runtimeID:     runtimeID,
		runtime:       runtime,
		runtimePrvKey: runtimePrvKey}
}

func TestAddProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, in)
	addedProcess1, err := client.PublishProcessSpec(processSpec1, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess2, err := client.PublishProcessSpec(processSpec2, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, processSpec2.TargetColonyID, addedProcess2.ProcessSpec.TargetColonyID)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.runtimeID, env.colonyID, 100, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArrayEqual(processes, processesFromServer))

	server.Shutdown()
	<-done
}

func TestApproveRuntime(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	runtimeFromServer, err := client.GetRuntimeByID(env.runtimeID, env.colonyID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	err = client.ApproveRuntime(env.runtime, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntimeByID(env.runtimeID, env.colonyID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, runtimeFromServer.IsApproved())

	err = client.RejectRuntime(env.runtime, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	runtimeFromServer, err = client.GetRuntimeByID(env.runtimeID, env.colonyID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.False(t, runtimeFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess1, err := client.PublishProcessSpec(processSpec1, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess2, err := client.PublishProcessSpec(processSpec2, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(env.runtimeID, env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.AssignProcess(env.runtimeID, env.colonyID, env.runtimePrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestMarkSuccessful(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.PublishProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.runtimeID, env.colonyID, env.runtimePrvKey)
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

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.PublishProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.runtimeID, env.colonyID, env.runtimePrvKey)
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

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_runtime", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.PublishProcessSpec(processSpec, env.runtimePrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.runtimeID, env.colonyID, env.runtimePrvKey)
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
