package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
	"colonies/pkg/security"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func PrepareTests(t *testing.T, rootPassword string) (*ColoniesServer, chan bool) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	debug := true
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
	colonyAdded, err := client.AddColony(colony, rootPassword)
	assert.Nil(t, err)

	colonyFromServer, err := client.GetColonyByID(colonyID, prvKey)
	assert.Nil(t, err)
	assert.Equal(t, colony.ID(), colonyFromServer.ID())
	assert.Equal(t, colony.Name(), colonyFromServer.Name())
	assert.Equal(t, colony.ID(), colonyAdded.ID())
	assert.Equal(t, colony.Name(), colonyAdded.Name())

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
	_, err = client.AddColony(colony1, rootPassword)
	assert.Nil(t, err)

	prvKey2, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID2, err := security.GenerateID(prvKey2)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colonyID2, "test_colony_name")
	_, err = client.AddColony(colony2, rootPassword)
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

	_, err = client.AddColony(colony, rootPassword)
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
	addedComputer, err := client.AddComputer(computer, colonyPrvKey)
	assert.Nil(t, err)

	computerFromServer, err := client.GetComputerByID(computerID, colonyID, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, computerFromServer)
	assert.Equal(t, computerID, computerFromServer.ID())
	assert.Equal(t, computerID, addedComputer.ID())

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

	_, err = client.AddColony(colony, rootPassword)
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
	_, err = client.AddComputer(computer1, colonyPrvKey)
	assert.Nil(t, err)

	// Create a computer2
	computer2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	computer2ID, err := security.GenerateID(computer2PrvKey)
	assert.Nil(t, err)

	name = "test_computer 2"

	computer2 := core.CreateComputer(computer2ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddComputer(computer2, colonyPrvKey)
	assert.Nil(t, err)

	computers, err := client.GetComputersByColonyID(colonyID, colonyPrvKey)
	assert.Nil(t, err)
	assert.Len(t, computers, 2)

	server.Shutdown()
	<-done
}

type clientTestEnv struct {
	colonyID       string
	colony         *core.Colony
	colonyPrvKey   string
	computerID     string
	computer       *core.Computer
	computerPrvKey string
}

func createTestEnv(t *testing.T, rootPassword string) *clientTestEnv {
	// Create a Colony
	colonyPrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := security.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, rootPassword)
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
	_, err = client.AddComputer(computer, colonyPrvKey)
	assert.Nil(t, err)

	return &clientTestEnv{colonyID: colonyID,
		colony:         colony,
		colonyPrvKey:   colonyPrvKey,
		computerID:     computerID,
		computer:       computer,
		computerPrvKey: computerPrvKey}
}

func TestAddProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	process1 := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	var attributes1 []*core.Attribute
	attributes1 = append(attributes1, core.CreateAttribute(process1.ID(), core.IN, "test_key_1", "test_value_2"))
	process1.SetInAttributes(attributes1)
	_, err := client.AddProcess(process1, env.computerPrvKey)
	assert.Nil(t, err)

	process2 := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	addedProcess2, err := client.AddProcess(process2, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process2.ID(), addedProcess2.ID())

	processes, err := client.GetWaitingProcesses(env.computerID, env.colonyID, 100, env.computerPrvKey)
	assert.Nil(t, err)

	counter := 0
	for _, processFromServer := range processes {
		if processFromServer.ID() == process1.ID() {
			counter++
		}
		if processFromServer.ID() == process2.ID() {
			counter++
		}
	}
	assert.Equal(t, 2, counter)

	server.Shutdown()
	<-done
}

func TestApproveComputer(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	computerFromServer, err := client.GetComputerByID(env.computerID, env.colonyID, env.colonyPrvKey)
	assert.Nil(t, err)
	assert.False(t, computerFromServer.IsApproved())

	err = client.ApproveComputer(env.computer, env.colonyPrvKey)
	assert.Nil(t, err)

	computerFromServer, err = client.GetComputerByID(env.computerID, env.colonyID, env.colonyPrvKey)
	assert.Nil(t, err)
	assert.True(t, computerFromServer.IsApproved())

	err = client.RejectComputer(env.computer, env.colonyPrvKey)
	assert.Nil(t, err)

	computerFromServer, err = client.GetComputerByID(env.computerID, env.colonyID, env.colonyPrvKey)
	assert.Nil(t, err)
	assert.False(t, computerFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	process1 := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	_, err := client.AddProcess(process1, env.computerPrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	_, err = client.AddProcess(process2, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process1.ID(), assignedProcess.ID())

	assignedProcess, err = client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, process2.ID(), assignedProcess.ID())

	server.Shutdown()
	<-done
}

func TestMarkSuccessful(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	process := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	_, err := client.AddProcess(process, env.computerPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcessByID(process.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.PENDING, processFromServer.Status())

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status())

	err = client.MarkSuccessful(assignedProcess, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.Status())

	server.Shutdown()
	<-done
}

func TestMarkFailed(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	process := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	_, err := client.AddProcess(process, env.computerPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcessByID(process.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.PENDING, processFromServer.Status())

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status())

	err = client.MarkFailed(assignedProcess, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.FAILED, assignedProcessFromServer.Status())

	server.Shutdown()
	<-done
}

func TestAddAttributes(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	process := core.CreateProcess(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	_, err := client.AddProcess(process, env.computerPrvKey)
	assert.Nil(t, err)

	processFromServer, err := client.GetProcessByID(process.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.PENDING, processFromServer.Status())

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID(), core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID(), addedAttribute.ID())

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID(), env.colonyID, env.computerPrvKey)
	assert.Equal(t, "helloworld", assignedProcessFromServer.Out()["result"])

	attributeFromServer, err := client.GetAttribute(attribute.ID(), process.ID(), env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID(), attributeFromServer.ID())

	server.Shutdown()
	<-done
}
