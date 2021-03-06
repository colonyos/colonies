package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestAddColonySecurity(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()

	privateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := crypto.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	_, err = client.AddColony(colony, "invalid_api_key")
	assert.NotNilf(t, err, "it should be possible to create a colony without correct api key")

	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteColonySecurity(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	crypto := crypto.CreateCrypto()

	privateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	colonyID, err := crypto.GenerateID(privateKey)
	assert.Nil(t, err)

	colony := core.CreateColony(colonyID, "test_colony_name")

	invalidPrivateKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	err = client.DeleteColony(colony.ID, invalidPrivateKey)
	assert.NotNil(t, err)

	err = client.DeleteColony(colony.ID, serverPrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetColoniesSecurity(t *testing.T) {
	_, client, server, serverPrvKey, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now, try to get colonies info using an invalid api
	_, err := client.GetColonies(core.GenerateRandomID())
	assert.NotNil(t, err) // Should not work

	// Now, try to get colonies info using an invalid api
	_, err = client.GetColonies(serverPrvKey)
	assert.Nil(t, err) // Should work

	server.Shutdown()
	<-done
}

func TestGetColonyByIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	// Now, try to get colony1 info using colony2 credentials
	_, err := client.GetColonyByID(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	// Now, try to get colony1 info using colony1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should work, cannot use colony1PrvKey as credential

	// Now, try to get colony1 info using runtime1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	// Now, try to get colony1 info using runtime1 credentials
	_, err = client.GetColonyByID(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}

func TestGetProcessStatSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   runtime1 is member of colony1
	//   runtime2 is member of colony2

	_, err := client.ColonyStatistics(env.colony1ID, env.runtime2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.ColonyStatistics(env.colony2ID, env.runtime1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.ColonyStatistics(env.colony1ID, env.runtime1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.ColonyStatistics(env.colony2ID, env.runtime2PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.ColonyStatistics(env.colony1ID, env.colony1PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.ColonyStatistics(env.colony2ID, env.colony2PrvKey)
	assert.Nil(t, err) // Should work

	_, err = client.ColonyStatistics(env.colony1ID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	server.Shutdown()
	<-done
}
