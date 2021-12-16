package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PrepareTests(t *testing.T, apiKey string) (*APIServer, chan bool) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	controller := CreateColoniesController(db)
	apiServer := CreateAPIServer(controller, 8080, apiKey, "../../cert/key.pem", "../../cert/cert.pem")
	done := make(chan bool)

	go func() {
		apiServer.ServeForever()
		done <- true
	}()

	return apiServer, done
}

func TestAddColony(t *testing.T) {
	apiKey := "testapikey"
	apiServer, done := PrepareTests(t, apiKey)

	privateKey, err := client.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := client.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, "invalid_api_key")
	assert.NotNilf(t, err, "it should be possible to create a colony without correct api key")

	err = client.AddColony(colony, apiKey)
	assert.Nil(t, err)

	apiServer.Shutdown()
	<-done
}

func TestAddWorker(t *testing.T) {
	apiKey := "testapikey"
	apiServer, done := PrepareTests(t, apiKey)

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
	fmt.Println(err)
	assert.Nil(t, err)

	apiServer.Shutdown()
	<-done
}
