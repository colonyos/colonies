package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
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

	privateKey, err := client.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := client.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, "invalid_api_key")
	assert.NotNilf(t, err, "it should be possible to create a colony without correct api key")

	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColony(colonyID, privateKey)
	assert.Nil(t, err)

	assert.Equal(t, colony.ID(), colonyFromServer.ID())
	assert.Equal(t, colony.Name(), colonyFromServer.Name())

	server.Shutdown()
	<-done
}

func TestGetColonies(t *testing.T) {
	apiKey := "testapikey"
	server, done := PrepareTests(t, apiKey)

	privateKey1, err := client.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID1, err := client.GenerateID(privateKey1)
	assert.Nil(t, err)
	colony1 := core.CreateColony(colonyID1, "test_colony_name")
	err = client.AddColony(colony1, apiKey)
	assert.Nil(t, err)

	privateKey2, err := client.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID2, err := client.GenerateID(privateKey2)
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
	colonyPrivateKey, err := client.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := client.GenerateID(colonyPrivateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	// Create a worker
	workerPrivateKey, err := client.GeneratePrivateKey()
	assert.Nil(t, err)

	workerID, err := client.GenerateID(workerPrivateKey)
	assert.Nil(t, err)

	name := "test_worker"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	worker := core.CreateWorker(workerID, name, colonyID, cpu, cores, mem, gpu, gpus)

	err = client.AddWorker(worker, colonyPrivateKey)
	assert.Nil(t, err)

	workerFromServer, err := client.GetWorker(workerID, colonyID, colonyPrivateKey)
	assert.Nil(t, err)
	assert.Equal(t, workerID, workerFromServer.ID())

	server.Shutdown()
	<-done
}
