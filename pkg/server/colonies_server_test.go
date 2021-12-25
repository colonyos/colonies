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

func TestAddComputer(t *testing.T) {
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
	addedComputer, err := client.AddComputer(computer, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, computer.Equals(addedComputer))

	computerFromServer, err := client.GetComputerByID(computerID, colonyID, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.NotNil(t, computerFromServer)
	assert.True(t, computer.Equals(computerFromServer))

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

	_, err = client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a Computer
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
	_, err = client.AddComputer(computer1, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a Computer
	computer2PrvKey, err := security.GeneratePrivateKey()
	assert.Nil(t, err)
	computer2ID, err := security.GenerateID(computer2PrvKey)
	assert.Nil(t, err)

	name = "test_computer 2"
	computer2 := core.CreateComputer(computer2ID, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddComputer(computer2, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	var computers []*core.Computer
	computers = append(computers, computer1)
	computers = append(computers, computer2)

	computersFromServer, err := client.GetComputersByColonyID(colonyID, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, core.IsComputerArraysEqual(computers, computersFromServer))

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

	_, err = client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
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
	_, err = client.AddComputer(computer, colonyPrvKey, TESTHOST, TESTPORT)
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

	in := make(map[string]string)
	in["test_key_1"] = "test_value_1"
	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, in)
	addedProcess1, err := client.PublishProcessSpec(processSpec1, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess2, err := client.PublishProcessSpec(processSpec2, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, processSpec2.TargetColonyID, addedProcess2.TargetColonyID)

	var processes []*core.Process
	processes = append(processes, addedProcess1)
	processes = append(processes, addedProcess2)

	processesFromServer, err := client.GetWaitingProcesses(env.computerID, env.colonyID, 100, env.computerPrvKey)
	assert.Nil(t, err)
	assert.True(t, core.IsProcessArrayEqual(processes, processesFromServer))

	server.Shutdown()
	<-done
}

func TestApproveComputer(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	computerFromServer, err := client.GetComputerByID(env.computerID, env.colonyID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.False(t, computerFromServer.IsApproved())

	err = client.ApproveComputer(env.computer, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	computerFromServer, err = client.GetComputerByID(env.computerID, env.colonyID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.True(t, computerFromServer.IsApproved())

	err = client.RejectComputer(env.computer, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	computerFromServer, err = client.GetComputerByID(env.computerID, env.colonyID, env.colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.False(t, computerFromServer.IsApproved())

	server.Shutdown()
	<-done
}

func TestAssignProcess(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec1 := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess1, err := client.PublishProcessSpec(processSpec1, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	processSpec2 := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess2, err := client.PublishProcessSpec(processSpec2, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess1.ID, assignedProcess.ID)

	assignedProcess, err = client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedProcess2.ID, assignedProcess.ID)

	server.Shutdown()
	<-done
}

func TestMarkSuccessful(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.PublishProcessSpec(processSpec, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status)

	err = client.MarkSuccessful(assignedProcess, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID, env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.SUCCESS, assignedProcessFromServer.Status)

	server.Shutdown()
	<-done
}

func TestMarkFailed(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.PublishProcessSpec(processSpec, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.RUNNING, assignedProcessFromServer.Status)

	err = client.MarkFailed(assignedProcess, env.computerPrvKey)
	assert.Nil(t, err)

	assignedProcessFromServer, err = client.GetProcessByID(assignedProcess.ID, env.colonyID, env.computerPrvKey)
	assert.Equal(t, core.FAILED, assignedProcessFromServer.Status)

	server.Shutdown()
	<-done
}

func TestAddAttributes(t *testing.T) {
	rootPassword := "password"
	server, done := PrepareTests(t, rootPassword)
	env := createTestEnv(t, rootPassword)

	processSpec := core.CreateProcessSpec(env.colonyID, []string{}, "test_computer", -1, 3, 1000, 10, 1, make(map[string]string))
	addedProcess, err := client.PublishProcessSpec(processSpec, env.computerPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)
	assert.Equal(t, core.PENDING, addedProcess.Status)

	assignedProcess, err := client.AssignProcess(env.computerID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(assignedProcess.ID, core.OUT, "result", "helloworld")
	addedAttribute, err := client.AddAttribute(attribute, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, addedAttribute.ID)

	assignedProcessFromServer, err := client.GetProcessByID(assignedProcess.ID, env.colonyID, env.computerPrvKey)

	out := make(map[string]string)
	for _, attribute := range assignedProcessFromServer.Attributes {
		out[attribute.Key] = attribute.Value
	}

	assert.Equal(t, "helloworld", out["result"])

	attributeFromServer, err := client.GetAttribute(attribute.ID, addedProcess.ID, env.colonyID, env.computerPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, attribute.ID, attributeFromServer.ID)

	server.Shutdown()
	<-done
}
