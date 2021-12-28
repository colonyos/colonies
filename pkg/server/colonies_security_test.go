package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO: Add more tests

type testEnv struct {
	colony1PrvKey  string
	colony1ID      string
	colony2PrvKey  string
	colony2ID      string
	runtime1PrvKey string
	runtime1ID     string
	runtime2PrvKey string
	runtime2ID     string
}

func generateRuntime(t *testing.T, colonyID string) (*core.Runtime, string, string) {
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

	return core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus), runtimeID, runtimePrvKey
}

func setupTestEnvironment(t *testing.T) (*testEnv, *ColoniesServer, chan bool) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	// Create a colony
	colony1PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colony1ID, err := security.GenerateID(colony1PrvKey)
	assert.Nil(t, err)
	colony1 := core.CreateColony(colony1ID, "test_colony_name")
	_, err = client.AddColony(colony1, apiKey, TESTHOST, TESTPORT)

	// Create a colony
	colony2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colony2ID, err := security.GenerateID(colony2PrvKey)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colony2ID, "test_colony_name")
	_, err = client.AddColony(colony2, apiKey, TESTHOST, TESTPORT)

	// Create a runtime
	runtime1, runtime1ID, runtime1PrvKey := generateRuntime(t, colony1ID)
	_, err = client.AddRuntime(runtime1, colony1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a runtime
	runtime2, runtime2ID, runtime2PrvKey := generateRuntime(t, colony2ID)
	_, err = client.AddRuntime(runtime2, colony2PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	env := &testEnv{colony1PrvKey: colony1PrvKey, colony1ID: colony1ID, colony2PrvKey: colony2PrvKey, colony2ID: colony2ID, runtime1PrvKey: runtime1PrvKey, runtime1ID: runtime1ID, runtime2PrvKey: runtime2PrvKey, runtime2ID: runtime2ID}

	return env, server, done
}

func TestAddColonySecurity(t *testing.T) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	privateKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, "invalid_api_key", TESTHOST, TESTPORT)
	assert.NotNilf(t, err, "it should be possible to create a colony without correct api key")

	_, err = client.AddColony(colony, apiKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetColoniesSecurity(t *testing.T) {
	_, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now, try to get colonies info using an invalid api
	_, err := client.GetColonies("invalid-api-key", TESTHOST, TESTPORT)
	assert.NotNil(t, err) // Should not work

	// Now, try to get colonies info using an invalid api
	_, err = client.GetColonies("testapikey", TESTHOST, TESTPORT)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetColonyByIDSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now, try to get colony1 info using colony2 credentials
	_, err := client.GetColonyByID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to get colony1 info using colony1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

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
	env, server, done := setupTestEnvironment(t)
	runtime3, _, _ := generateRuntime(t, env.colony1ID)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2
	//   runtime3 is bound to colony1, but not yet a member

	// Now, try to add runtime 3 to colony1 using colony 2 credentials
	_, err := client.AddRuntime(runtime3, env.colony2PrvKey, TESTHOST, TESTPORT)
	assert.NotNil(t, err) // Should not work

	// Now, try to add runtime 3 to colony1 using colony 1 credentials
	_, err = client.AddRuntime(runtime3, env.colony1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetRuntimeByIDSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now try to access runtime1 using credentials of runtime2
	_, err := client.GetRuntimeByID(env.runtime1ID, env.colony1ID, env.runtime2PrvKey, TESTHOST, TESTPORT)
	assert.NotNil(t, err) // Should not work

	// Now try to access runtime1 using runtime1 credentials
	_, err = client.GetRuntimeByID(env.runtime1ID, env.colony1ID, env.runtime1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credentials
	_, err = client.GetRuntimeByID(env.runtime1ID, env.colony1ID, env.colony1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credentials
	_, err = client.GetRuntimeByID(env.runtime1ID, env.colony1ID, env.colony2PrvKey, TESTHOST, TESTPORT)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestGetRuntimesByColonySecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now try to access runtime1 using credentials of runtime2
	_, err := client.GetRuntimesByColonyID(env.colony1ID, env.runtime2PrvKey, TESTHOST, TESTPORT)
	assert.NotNil(t, err) // Should not work

	// Now try to access runtime1 using runtime1 credentials
	_, err = client.GetRuntimesByColonyID(env.colony1ID, env.runtime1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credentials
	_, err = client.GetRuntimesByColonyID(env.colony1ID, env.colony1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err) // Should work

	// Now try to access runtime1 using colony1 credentials
	_, err = client.GetRuntimesByColonyID(env.colony1ID, env.colony2PrvKey, TESTHOST, TESTPORT)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestAssignProcessSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	processSpec1 := core.CreateProcessSpec(env.colony1ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	_, err := client.SubmitProcessSpec(processSpec1, env.runtime1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(env.colony2ID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, make(map[string]string))
	_, err = client.SubmitProcessSpec(processSpec2, env.runtime2PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Now try to assign a process from colony2 using runtime1 credentials
	_, err = client.AssignProcess(env.runtime2ID, env.colony1ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to assign a process from colony2 using runtime1 credentials
	_, err = client.AssignProcess(env.runtime1ID, env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.AssignProcess(env.runtime1ID, env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work, only runtimes are allowed

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.AssignProcess(env.runtime1ID, env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work, only runtimes are allowed, also invalid credentials are used

	server.Shutdown()
	<-done
}
