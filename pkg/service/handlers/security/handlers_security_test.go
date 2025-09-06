package security_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeUserIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   user1 is member of colony1
	//   user2 is member of colony1
	//   user3 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.Colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.Colony1PrvKey)
	assert.Nil(t, err)

	user3, user3PrvKey, err := utils.CreateTestUserWithKey(env.Colony2Name, "test_user3")
	assert.Nil(t, err)
	_, err = client.AddUser(user3, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newUserPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newUserID, err := crypto.GenerateID(newUserPrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, user2PrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, user1PrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, user3PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, env.Executor1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, env.Executor2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, env.Colony1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeUserID(env.Colony1Name, newUserID, env.Colony2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeExecutorIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   user1 is member of colony1
	//   user2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.Colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.Colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newExecutorPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newExecutorID, err := crypto.GenerateID(newExecutorPrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(env.Colony1Name, newExecutorID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony1Name, newExecutorID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony2Name, newExecutorID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony2Name, newExecutorID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony2Name, newExecutorID, env.Colony1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony2Name, newExecutorID, env.Colony2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony1Name, newExecutorID, env.Executor2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeExecutorID(env.Colony1Name, newExecutorID, env.Executor1PrvKey)
	assert.Nil(t, err)

	err = client.ChangeExecutorID(env.Colony1Name, newExecutorID, env.Executor1PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeColonyIDSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2
	//   user1 is member of colony1
	//   user2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.Colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.Colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.Colony2PrvKey)
	assert.Nil(t, err)

	// Change Id
	crypto := crypto.CreateCrypto()
	newColonyPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newColonyID, err := crypto.GenerateID(newColonyPrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, user1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, user2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, env.Executor1PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, env.Executor2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, "blabla")
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, env.Colony2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, env.Colony1PrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(env.Colony2Name, newColonyID, env.Colony2PrvKey)
	assert.Nil(t, err)

	err = client.ChangeColonyID(env.Colony2Name, newColonyID, env.Colony2PrvKey)
	assert.NotNil(t, err)

	err = client.ChangeColonyID(env.Colony1Name, newColonyID, env.Colony1PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestChangeServerIDSecurity(t *testing.T) {
	client, server, serverPrvKey, done := server.PrepareTests(t)

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
