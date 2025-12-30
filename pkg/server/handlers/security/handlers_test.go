package security_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeUserID(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	user, userPrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "test_user")
	assert.Nil(t, err)

	addedUser, err := client.AddUser(user, env.ColonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	// Check if the user can use the API
	users, err := client.GetUsers(env.ColonyName, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	// Change Id
	crypto := crypto.CreateCrypto()
	newUserPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newUserID, err := crypto.GenerateID(newUserPrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.ColonyName, newUserID, userPrvKey)
	assert.Nil(t, err)

	// Check if the user can use the API with the new private key
	users, err = client.GetUsers(env.ColonyName, newUserPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	// Check if the user can use the API with the OLD private key, should not work
	_, err = client.GetUsers(env.ColonyName, userPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeExecutorID(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	executors, err := client.GetExecutors(env.ColonyName, executorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executors)

	// Change Id
	crypto := crypto.CreateCrypto()
	newExecutorPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newExecutorID, err := crypto.GenerateID(newExecutorPrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(env.ColonyName, newExecutorID, executorPrvKey)
	assert.Nil(t, err)

	// Check if the user can use the API with the new private key
	executors, err = client.GetExecutors(env.ColonyName, newExecutorPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, executors)

	// Check if the executor can use the API with the OLD private key, should not work
	_, err = client.GetExecutors(env.ColonyName, executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeColonyID(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

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
	client, server, serverPrvKey, done := server.PrepareTests(t)

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

// TestChangeUserIDUnauthorized tests that non-members cannot change user ID
func TestChangeUserIDUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	user, userPrvKey, err := utils.CreateTestUserWithKey(colony1.Name, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(colony2.Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to change user ID in colony1 with user2's key from colony2
	cryptoLib := crypto.CreateCrypto()
	newUserID, err := cryptoLib.GenerateID(userPrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(colony1.Name, newUserID, user2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestChangeExecutorIDUnauthorized tests that non-members cannot change executor ID
func TestChangeExecutorIDUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, _, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to change executor ID in colony1 with executor2's key from colony2
	cryptoLib := crypto.CreateCrypto()
	newExecutorPrvKey, err := cryptoLib.GeneratePrivateKey()
	assert.Nil(t, err)
	newExecutorID, err := cryptoLib.GenerateID(newExecutorPrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(colony1.Name, newExecutorID, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestChangeColonyIDUnauthorized tests that non-owners cannot change colony ID
func TestChangeColonyIDUnauthorized(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	// Try to change colony ID with executor key (not colony owner)
	cryptoLib := crypto.CreateCrypto()
	newColonyPrvKey, err := cryptoLib.GeneratePrivateKey()
	assert.Nil(t, err)
	newColonyID, err := cryptoLib.GenerateID(newColonyPrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(colony.Name, newColonyID, executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestChangeServerIDUnauthorized tests that non-server-owners cannot change server ID
func TestChangeServerIDUnauthorized(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to change server ID with colony key (not server owner)
	cryptoLib := crypto.CreateCrypto()
	newServerPrvKey, err := cryptoLib.GeneratePrivateKey()
	assert.Nil(t, err)
	newServerID, err := cryptoLib.GenerateID(newServerPrvKey)
	assert.Nil(t, err)

	err = client.ChangeServerID(newServerID, colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestChangeUserIDInvalidLength tests changing user ID with invalid length
func TestChangeUserIDInvalidLength(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	user, userPrvKey, err := utils.CreateTestUserWithKey(env.ColonyName, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to change user ID with invalid length
	err = client.ChangeUserID(env.ColonyName, "short_id", userPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestChangeExecutorIDInvalidLength tests changing executor ID with invalid length
func TestChangeExecutorIDInvalidLength(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to change executor ID with invalid length
	err := client.ChangeExecutorID(env.ColonyName, "short_id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestChangeColonyIDInvalidLength tests changing colony ID with invalid length
func TestChangeColonyIDInvalidLength(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to change colony ID with invalid length
	err = client.ChangeColonyID(colony.Name, "short_id", colonyPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
