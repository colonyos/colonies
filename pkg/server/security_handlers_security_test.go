package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeUserIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   user1 is member of colony1
	//   user2 is member of colony1
	//   user3 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.colony1PrvKey)
	assert.Nil(t, err)

	user3, user3PrvKey, err := utils.CreateTestUserWithKey(env.colony2Name, "test_user3")
	assert.Nil(t, err)
	_, err = client.AddUser(user3, env.colony2PrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newUserPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newUserID, err := crypto.GenerateID(newUserPrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, user2PrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, user1PrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, user3PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, env.executor1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, env.executor2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, env.colony1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.colony1Name, newUserID, env.colony2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeExecutorIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   user1 is member of colony1
	//   user2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.colony2PrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newExecutorPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newExecutorID, err := crypto.GenerateID(newExecutorPrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(env.colony1Name, newExecutorID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony1Name, newExecutorID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony2Name, newExecutorID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony2Name, newExecutorID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony2Name, newExecutorID, env.colony1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony2Name, newExecutorID, env.colony2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony1Name, newExecutorID, env.executor2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.colony1Name, newExecutorID, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(env.colony1Name, newExecutorID, env.executor1PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeColonyIDSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   user1 is member of colony1
	//   user2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.colony2PrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newColonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newColonyID, err := crypto.GenerateID(newColonyPrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, env.executor1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, env.executor2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, "blabla")
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, env.colony2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(env.colony2Name, newColonyID, env.colony2PrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(env.colony2Name, newColonyID, env.colony2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.colony1Name, newColonyID, env.colony1PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeServerIDSecurity(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, userPrvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newServerPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newServerID, err := crypto.GenerateID(newServerPrvKey)
	assert.Nil(t, err)

	err = client.ChangeServerID(newServerID, executorPrvKey)
	assert.NotNil(t, err)

	err = client.ChangeServerID(newServerID, userPrvKey)
	assert.NotNil(t, err)

	err = client.ChangeServerID(newServerID, colonyPrvKey)
	assert.NotNil(t, err)

	err = client.ChangeServerID(newServerID, serverPrvKey)
	assert.Nil(t, err)

	err = client.ChangeServerID(newServerID, serverPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
