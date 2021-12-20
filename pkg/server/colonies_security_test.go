package server

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testEnv struct {
	colony1PrvKey   string
	colony1ID       string
	colony2PrvKey   string
	colony2ID       string
	computer1PrvKey string
	computer1ID     string
	computer2PrvKey string
	computer2ID     string
}

func generateComputer(t *testing.T, colonyID string) (*core.Computer, string, string) {
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

	return core.CreateComputer(computerID, name, colonyID, cpu, cores, mem, gpu, gpus), computerID, computerPrvKey
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

	// Create a computer
	computer1, computer1ID, computer1PrvKey := generateComputer(t, colony1ID)
	err = client.AddComputer(computer1, colony1PrvKey)
	assert.Nil(t, err)

	// Create a computer
	computer2, computer2ID, computer2PrvKey := generateComputer(t, colony2ID)
	err = client.AddComputer(computer2, colony2PrvKey)
	assert.Nil(t, err)

	env := &testEnv{colony1PrvKey: colony1PrvKey, colony1ID: colony1ID, colony2PrvKey: colony2PrvKey, colony2ID: colony2ID, computer1PrvKey: computer1PrvKey, computer1ID: computer1ID, computer2PrvKey: computer2PrvKey, computer2ID: computer2ID}

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
	//   computer1 is member of colony1
	//   computer2 is member of colony2

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
	//   computer1 is member of colony1
	//   computer2 is member of colony2

	// Now, try to get colony1 info using colony2 credentials
	_, err := client.GetColonyByID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to get colony1 info using colony1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	// Now, try to get colony1 info using computer1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.computer1PrvKey)
	assert.Nil(t, err) // Should work

	// Now, try to get colony1 info using computer1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.computer2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestAddComputerSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)
	computer3, _, _ := generateComputer(t, env.colony1ID)

	// The setup looks like this:
	//   computer1 is member of colony1
	//   computer2 is member of colony2
	//   computer3 is bound to colony1, but not yet a member

	// Now, try to add computer 3 to colony1 using colony 2 credentials
	err := client.AddComputer(computer3, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to add computer 3 to colony1 using colony 1 credentials
	err = client.AddComputer(computer3, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetComputerByIDSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   computer1 is member of colony1
	//   computer2 is member of colony2

	// Now try to access computer1 using credentials of computer2
	_, err := client.GetComputerByID(env.computer1ID, env.colony1ID, env.computer2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access computer1 using computer1 credentials
	_, err = client.GetComputerByID(env.computer1ID, env.colony1ID, env.computer1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access computer1 using colony1 credentials
	_, err = client.GetComputerByID(env.computer1ID, env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access computer1 using colony1 credentials
	_, err = client.GetComputerByID(env.computer1ID, env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestGetComputersByColonySecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   computer1 is member of colony1
	//   computer2 is member of colony2

	// Now try to access computer1 using credentials of computer2
	_, err := client.GetComputersByColonyID(env.colony1ID, env.computer2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to access computer1 using computer1 credentials
	_, err = client.GetComputersByColonyID(env.colony1ID, env.computer1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access computer1 using colony1 credentials
	_, err = client.GetComputersByColonyID(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to access computer1 using colony1 credentials
	_, err = client.GetComputersByColonyID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestAssignProcessSecurity(t *testing.T) {
	env, server, done := setupTestEnvironment(t)

	// The setup looks like this:
	//   computer1 is member of colony1
	//   computer2 is member of colony2

	process1 := core.CreateProcess(env.colony1ID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	err := client.AddProcess(process1, env.computer1PrvKey)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	process2 := core.CreateProcess(env.colony2ID, []string{}, "test_computer", -1, 3, 1000, 10, 1)
	err = client.AddProcess(process2, env.computer2PrvKey)
	assert.Nil(t, err)

	// Now try to assign a process from colony2 using computer1 credentials
	_, err = client.AssignProcess(env.computer2ID, env.colony1ID, env.computer1PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now try to assign a process from colony2 using computer1 credentials
	_, err = client.AssignProcess(env.computer1ID, env.colony1ID, env.computer1PrvKey)
	assert.Nil(t, err) // Should work

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.AssignProcess(env.computer1ID, env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work, only computers are allowed

	// Now try to assign a process from colony2 using colony1 credentials
	_, err = client.AssignProcess(env.computer1ID, env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work, only computers are allowed, also invalid credentials are used

	server.Shutdown()
	<-done
}
