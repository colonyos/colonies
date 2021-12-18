package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEnv struct {
	colony1PrvKey string
	colony1ID     string
	colony2PrvKey string
	colony2ID     string
	worker1PrvKey string
	worker1ID     string
	worker2PrvKey string
	worker2ID     string
}

func generateWorker(t *testing.T, colonyID string) (*core.Worker, string, string) {
	workerPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	workerID, err := security.GenerateID(workerPrvKey)
	assert.Nil(t, err)

	name := "test_worker"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	return core.CreateWorker(workerID, name, colonyID, cpu, cores, mem, gpu, gpus), workerID, workerPrvKey
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
	err = client.AddColony(colony1, apiKey)

	// Create a colony
	colony2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colony2ID, err := security.GenerateID(colony2PrvKey)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colony2ID, "test_colony_name")
	err = client.AddColony(colony2, apiKey)

	// Create a worker
	worker1, worker1ID, worker1PrvKey := generateWorker(t, colony1ID)
	err = client.AddWorker(worker1, colony1PrvKey)
	assert.Nil(t, err)

	// Create a worker
	worker2, worker2ID, worker2PrvKey := generateWorker(t, colony2ID)
	err = client.AddWorker(worker2, colony2PrvKey)
	assert.Nil(t, err)

	env := &testEnv{colony1PrvKey: colony1PrvKey, colony1ID: colony1ID, colony2PrvKey: colony2PrvKey, colony2ID: colony2ID, worker1PrvKey: worker1PrvKey, worker1ID: worker1ID, worker2PrvKey: worker2PrvKey, worker2ID: worker2ID}

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

	err = client.AddColony(colony, "invalid_api_key")
	assert.NotNilf(t, err, "it should be possible to create a colony without correct api key")

	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetColoniesSecurity(t *testing.T) {
	_, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   worker1 is member of colony1
	//   worker2 is member of colony2

	// Now, try to get colonies info using an invalid api
	_, err := client.GetColonies("invalid-api-key")
	assert.NotNil(t, err) // Should not work

	// Now, try to get colonies info using an invalid api
	_, err = client.GetColonies("testapikey")
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetColonyByIDSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   worker1 is member of colony1
	//   worker2 is member of colony2

	// Now, try to get colony1 info using colony2 credentials
	_, err := client.GetColonyByID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to get colony1 info using colony1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestAddWorkerSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)
	worker3, _, _ := generateWorker(t, env.colony1ID)

	// The setup looks like this:
	//   worker1 is member of colony1
	//   worker2 is member of colony2
	//   worker3 is bound to colony1, but not yet a member

	// Now, try to add worker 3 to colony1 using colony 2 credentials
	err := client.AddWorker(worker3, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to add worker 3 to colony1 using colony 1 credentials
	err = client.AddWorker(worker3, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetWorkerByIDSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   worker1 is member of colony1
	//   worker2 is member of colony2

	// Now try to access worker1 using credentials of worker2
	_, err := client.GetWorkerByID(env.worker1ID, env.colony1ID, env.worker2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access worker1 using worker1 credentials
	_, err = client.GetWorkerByID(env.worker1ID, env.colony1ID, env.worker1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access worker1 using colony1 credentials
	_, err = client.GetWorkerByID(env.worker1ID, env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access worker1 using colony1 credentials
	_, err = client.GetWorkerByID(env.worker1ID, env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestGetWorkersByColonySecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   worker1 is member of colony1
	//   worker2 is member of colony2

	// Now try to access worker1 using credentials of worker2
	_, err := client.GetWorkersByColonyID(env.colony1ID, env.worker2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access worker1 using worker1 credentials
	_, err = client.GetWorkersByColonyID(env.colony1ID, env.worker1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access worker1 using colony1 credentials
	_, err = client.GetWorkersByColonyID(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access worker1 using colony1 credentials
	_, err = client.GetWorkersByColonyID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}
