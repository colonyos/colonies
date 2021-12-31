package server

import (
	"colonies/internal/logging"
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/database/postgresql"
	"colonies/pkg/security/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEnv1 struct {
	colony1PrvKey  string
	colony1ID      string
	colony2PrvKey  string
	colony2ID      string
	runtime1PrvKey string
	runtime1ID     string
	runtime2PrvKey string
	runtime2ID     string
}

type testEnv2 struct {
	colonyID      string
	colony        *core.Colony
	colonyPrvKey  string
	runtimeID     string
	runtime       *core.Runtime
	runtimePrvKey string
}

func setupTestEnv1(t *testing.T) (*testEnv1, *ColoniesServer, chan bool) {
	rootPassword := "secretpassword"
	server, done := prepareTests(t, rootPassword)

	crypto := crypto.CreateCrypto()

	// Create a colony
	colony1PrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colony1ID, err := crypto.GenerateID(colony1PrvKey)
	assert.Nil(t, err)
	colony1 := core.CreateColony(colony1ID, "test_colony_name")
	_, err = client.AddColony(colony1, rootPassword, TESTHOST, TESTPORT)

	// Create a colony
	colony2PrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colony2ID, err := crypto.GenerateID(colony2PrvKey)
	assert.Nil(t, err)
	colony2 := core.CreateColony(colony2ID, "test_colony_name")
	_, err = client.AddColony(colony2, rootPassword, TESTHOST, TESTPORT)

	// Create a runtime
	runtime1, runtime1ID, runtime1PrvKey := generateRuntime(t, colony1ID)
	_, err = client.AddRuntime(runtime1, colony1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a runtime
	runtime2, runtime2ID, runtime2PrvKey := generateRuntime(t, colony2ID)
	_, err = client.AddRuntime(runtime2, colony2PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime1.ID, colony1PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime2.ID, colony2PrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	env := &testEnv1{colony1PrvKey: colony1PrvKey,
		colony1ID:      colony1ID,
		colony2PrvKey:  colony2PrvKey,
		colony2ID:      colony2ID,
		runtime1PrvKey: runtime1PrvKey,
		runtime1ID:     runtime1ID,
		runtime2PrvKey: runtime2PrvKey,
		runtime2ID:     runtime2ID}

	return env, server, done
}

func setupTestEnv2(t *testing.T) (*testEnv2, *ColoniesServer, chan bool) {
	rootPassword := "secretpassword"
	server, done := prepareTests(t, rootPassword)

	crypto := crypto.CreateCrypto()

	// Create a Colony
	colonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := crypto.GenerateID(colonyPrvKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, rootPassword, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	// Create a runtime
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	assert.Nil(t, err)

	name := "test_runtime_name"
	runtimeType := "test_runtime_type"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	runtime := core.CreateRuntime(runtimeID, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus)
	_, err = client.AddRuntime(runtime, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey, TESTHOST, TESTPORT)
	assert.Nil(t, err)

	env := &testEnv2{colonyID: colonyID,
		colony:        colony,
		colonyPrvKey:  colonyPrvKey,
		runtimeID:     runtimeID,
		runtime:       runtime,
		runtimePrvKey: runtimePrvKey}

	return env, server, done
}

func generateRuntime(t *testing.T, colonyID string) (*core.Runtime, string, string) {
	crypto := crypto.CreateCrypto()

	runtimePrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	runtimeID, err := crypto.GenerateID(runtimePrvKey)
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

func prepareTests(t *testing.T, rootPassword string) (*ColoniesServer, chan bool) {
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
