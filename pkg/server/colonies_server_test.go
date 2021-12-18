package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
	"colonies/pkg/security"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PrepareTests(t *testing.T, apiKey string) (*ColoniesServer, chan bool) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	server := CreateColoniesServer(db, 8080, apiKey, "../../cert/key.pem", "../../cert/cert.pem")
	done := make(chan bool)

	go func() {
		server.ServeForever()
		done <- true
	}()

	return server, done
}

func TestAddColony(t *testing.T) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	prvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(prvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")
	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colonyID, prvKey)
	assert.Nil(t, err)

	assert.Equal(t, colony.ID(), colonyFromServer.ID())
	assert.Equal(t, colony.Name(), colonyFromServer.Name())

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	prvKey1, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID1, err := security.GenerateID(prvKey1)
	assert.Nil(t, err)
	colony1 := core.CreateColony(colonyID1, "test_colony_name")
	err = client.AddColony(colony1, apiKey)
	assert.Nil(t, err)

	prvKey2, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID2, err := security.GenerateID(prvKey2)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colonyID2, "test_colony_name")
	err = client.AddColony(colony2, apiKey)
	assert.Nil(t, err)

	coloniesFromServer, err := client.GetColonies(apiKey)
	assert.Nil(t, err)

	counter := 0
	for _, colony := range coloniesFromServer {
		if colony.ID() == colonyID1 || colony.ID() == colonyID2 {
			counter++
		}
	}
	assert.Equal(t, 2, counter)

	server.Shutdown()
	<-done
}

func TestAddWorker(t *testing.T) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	// Create a worker
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

	worker := core.CreateWorker(workerID, name, colonyID, cpu, cores, mem, gpu, gpus)
	err = client.AddWorker(worker, colonyPrvKey)
	assert.Nil(t, err)

	workerFromServer, err := client.GetWorkerByID(workerID, colonyID, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, workerFromServer)
	assert.Equal(t, workerID, workerFromServer.ID())

	server.Shutdown()
	<-done
}

func TestGetWorkers(t *testing.T) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	// Create a worker 1
	worker1PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	worker1ID, err := security.GenerateID(worker1PrvKey)
	assert.Nil(t, err)

	name := "test_worker 1"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	worker1 := core.CreateWorker(worker1ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	err = client.AddWorker(worker1, colonyPrvKey)
	assert.Nil(t, err)

	// Create a worker2
	worker2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	worker2ID, err := security.GenerateID(worker2PrvKey)
	assert.Nil(t, err)

	name = "test_worker 2"

	worker2 := core.CreateWorker(worker2ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	err = client.AddWorker(worker2, colonyPrvKey)
	assert.Nil(t, err)

	workers, err := client.GetWorkersByColonyID(colonyID, colonyPrvKey)
	assert.Nil(t, err)
	assert.Len(t, workers, 2)

	server.Shutdown()
	<-done
}
