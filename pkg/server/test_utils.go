package server

import (
	"testing"

	"github.com/colonyos/colonies/internal/logging"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
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

func setupTestEnv1(t *testing.T) (*testEnv1, *client.ColoniesClient, *ColoniesServer, string, chan bool) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony1, colony1PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	runtime1, runtime1PrvKey, err := utils.CreateTestRuntimeWithKey(colony1.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime1, colony1PrvKey)
	assert.Nil(t, err)

	runtime2, runtime2PrvKey, err := utils.CreateTestRuntimeWithKey(colony2.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime2, colony2PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime1.ID, colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime2.ID, colony2PrvKey)
	assert.Nil(t, err)

	env := &testEnv1{colony1PrvKey: colony1PrvKey,
		colony1ID:      colony1.ID,
		colony2PrvKey:  colony2PrvKey,
		colony2ID:      colony2.ID,
		runtime1PrvKey: runtime1PrvKey,
		runtime1ID:     runtime1.ID,
		runtime2PrvKey: runtime2PrvKey,
		runtime2ID:     runtime2.ID}

	return env, client, server, serverPrvKey, done
}

func setupTestEnv2(t *testing.T) (*testEnv2, *client.ColoniesClient, *ColoniesServer, string, chan bool) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	_, err = client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	env := &testEnv2{colonyID: colony.ID,
		colony:        colony,
		colonyPrvKey:  colonyPrvKey,
		runtimeID:     runtime.ID,
		runtime:       runtime,
		runtimePrvKey: runtimePrvKey}

	return env, client, server, serverPrvKey, done
}

func prepareTests(t *testing.T) (*client.ColoniesClient, *ColoniesServer, string, chan bool) {
	client := client.CreateColoniesClient(TESTHOST, TESTPORT, true, true)

	debug := false
	if debug {
		logging.DisableDebug()
	}

	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	server := CreateColoniesServer(db, TESTPORT, serverID, true, "../../cert/key.pem", "../../cert/cert.pem", debug)
	done := make(chan bool)

	go func() {
		server.ServeForever()
		done <- true
	}()

	return client, server, serverPrvKey, done
}
