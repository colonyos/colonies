package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
	"colonies/pkg/security"
	"testing"

	"github.com/stretchr/testify/assert"
)

func PrepareTests(t *testing.T, rootPassword string) (*ColoniesServer, chan bool) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	server := CreateColoniesServer(db, 8080, rootPassword, "../../cert/key.pem", "../../cert/cert.pem")
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
	err = client.AddColony(colony, rootPassword)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colonyID, prvKey)
	assert.Nil(t, err)

	assert.Equal(t, colony.ID(), colonyFromServer.ID())
	assert.Equal(t, colony.Name(), colonyFromServer.Name())

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
	err = client.AddColony(colony1, rootPassword)
	assert.Nil(t, err)

	prvKey2, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID2, err := security.GenerateID(prvKey2)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colonyID2, "test_colony_name")
	err = client.AddColony(colony2, rootPassword)
	assert.Nil(t, err)

	coloniesFromServer, err := client.GetColonies(rootPassword)
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

func TestAddComputer(t *testing.T) {
	rootPassword := "testapikey"
	server, done := PrepareTests(t, rootPassword)

	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, rootPassword)
	assert.Nil(t, err)

	// Create a computer
	computerPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	computerID, err := security.GenerateID(computerPrvKey)
	assert.Nil(t, err)

	name := "test_computer"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	computer := core.CreateComputer(computerID, name, colonyID, cpu, cores, mem, gpu, gpus)
	err = client.AddComputer(computer, colonyPrvKey)
	assert.Nil(t, err)

	computerFromServer, err := client.GetComputerByID(computerID, colonyID, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, computerFromServer)
	assert.Equal(t, computerID, computerFromServer.ID())

	server.Shutdown()
	<-done
}

func TestGetComputers(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)

	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	err = client.AddColony(colony, rootPassword)
	assert.Nil(t, err)

	// Create a computer 1
	computer1PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	computer1ID, err := security.GenerateID(computer1PrvKey)
	assert.Nil(t, err)

	name := "test_computer 1"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	computer1 := core.CreateComputer(computer1ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	err = client.AddComputer(computer1, colonyPrvKey)
	assert.Nil(t, err)

	// Create a computer2
	computer2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	computer2ID, err := security.GenerateID(computer2PrvKey)
	assert.Nil(t, err)

	name = "test_computer 2"

	computer2 := core.CreateComputer(computer2ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	err = client.AddComputer(computer2, colonyPrvKey)
	assert.Nil(t, err)

	computers, err := client.GetComputersByColonyID(colonyID, colonyPrvKey)
	assert.Nil(t, err)
	assert.Len(t, computers, 2)

	server.Shutdown()
	<-done
}
