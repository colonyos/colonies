package user_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddUserSecurity(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user := utils.CreateTestUser(env.Colony1Name, "test_user")

	_, err := client.AddUser(user, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddUser(user, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddUser(user, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddUser(user, env.Colony1PrvKey)
	assert.Nil(t, err)

	s.Shutdown()
	<-done
}

func TestGetUsersSecurity(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.Colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.Colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.Colony2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.Colony1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.Colony1Name, user2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.Colony2Name, user1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.Colony2Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.Colony2Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should now work

	_, err = client.GetUsers(env.Colony2Name, user2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.Colony1Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.Colony1Name, user1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.Colony2Name, env.Executor2PrvKey)
	assert.Nil(t, err)

	s.Shutdown()
	<-done
}

func TestGetUserSecurity(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.Colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.Colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.Colony2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.Colony1Name, "test_user1", env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.Colony1Name, "test_user1", user2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.Colony2Name, "test_user2", user1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.Colony2Name, "test_user2", env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.Colony2Name, "test_user2", env.Colony1PrvKey)
	assert.NotNil(t, err) // Should now work

	_, err = client.GetUser(env.Colony2Name, "test_user2", user2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.Colony1Name, "test_user1", env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.Colony1Name, "test_user1", user1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.Colony2Name, "test_user2", env.Executor2PrvKey)
	assert.Nil(t, err)

	s.Shutdown()
	<-done
}

func TestRemoveUserSecurity(t *testing.T) {
	env, client, s, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.Colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.Colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.Colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.Colony2PrvKey)
	assert.Nil(t, err)

	err = client.RemoveUser(env.Colony1Name, "test_user1", env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.Colony1Name, "test_user1", env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.Colony1Name, "test_user1", user1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.Colony1Name, "test_user1", user2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.Colony1Name, "test_user1", env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.Colony1Name, "test_user1", env.Colony1PrvKey)
	assert.Nil(t, err)

	s.Shutdown()
	<-done
}