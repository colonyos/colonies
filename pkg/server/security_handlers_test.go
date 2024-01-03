package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeUserID(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	user, userPrvKey, err := utils.CreateTestUserWithKey(env.colonyName, "test_user")
	assert.Nil(t, err)

	addedUser, err := client.AddUser(user, env.colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	// Check if the user can use the API
	users, err := client.GetUsers(env.colonyName, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	// Change Id
	crypto := crypto.CreateCrypto()
	newUserPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newUserID, err := crypto.GenerateID(newUserPrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.colonyName, newUserID, userPrvKey)
	assert.Nil(t, err)

	// Check if the user can use the API with the new private key
	users, err = client.GetUsers(env.colonyName, newUserPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	// Check if the user can use the API with the OLD private key, should not work
	_, err = client.GetUsers(env.colonyName, userPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeExecutorID(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	executors, err := client.GetExecutors(env.colonyName, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executors)

	// Change Id
	crypto := crypto.CreateCrypto()
	newExecutorPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newExecutorID, err := crypto.GenerateID(newExecutorPrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(env.colonyName, newExecutorID, executorPrvKey)
	assert.Nil(t, err)

	// Check if the user can use the API with the new private key
	executors, err = client.GetExecutors(env.colonyName, newExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executors)

	// Check if the executor can use the API with the OLD private key, should not work
	_, err = client.GetExecutors(env.colonyName, executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeColonyID(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, _, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newColonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newColonyID, err := crypto.GenerateID(newColonyPrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(colony.Name, newColonyID, colonyPrvKey)
	assert.Nil(t, err)

	// Try to register a new executor with the new private key
	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, newColonyPrvKey)
	assert.Nil(t, err)

	// Just try to submit a function spec, to check if the executor is working after colony id change
	funcSpec2 := utils.CreateTestFunctionSpec(colony.Name)
	_, err = client.Submit(funcSpec2, executorPrvKey)

	// Try to register a new executor with the OLD private key, should not work
	executor, _, err = utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.NotNil(t, err) // Error

	server.Shutdown()
	<-done
}

func TestChangeServerID(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newServerPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newServerID, err := crypto.GenerateID(newServerPrvKey)
	assert.Nil(t, err)

	err = client.ChangeServerID(newServerID, serverPrvKey)
	assert.Nil(t, err)

	// Try to register a new colony with the new private key
	colony, colonyPrvKey, err = utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, newServerPrvKey)
	assert.Nil(t, err)

	// Try to register a new colony with the OLD private key, should not work
	_, _, err = utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.NotNil(t, err)

	// Just try to register a new user, to check if the server is working after server id change
	user, _, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)

	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	server.Shutdown()
	<-done
}
