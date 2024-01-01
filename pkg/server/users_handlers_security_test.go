package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddUserSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user := utils.CreateTestUser(env.colony1Name, "test_user")

	_, err := client.AddUser(user, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddUser(user, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddUser(user, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddUser(user, env.colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetUsersSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.colony2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.colony1Name, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.colony1Name, user2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.colony2Name, user1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.colony2Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUsers(env.colony2Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should now work

	_, err = client.GetUsers(env.colony2Name, user2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.colony1Name, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.colony1Name, user1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUsers(env.colony2Name, env.executor2PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetUserSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.colony2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.colony1Name, "test_user1", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.colony1Name, "test_user1", user2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.colony2Name, "test_user2", user1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.colony2Name, "test_user2", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetUser(env.colony2Name, "test_user2", env.colony1PrvKey)
	assert.NotNil(t, err) // Should now work

	_, err = client.GetUser(env.colony2Name, "test_user2", user2PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.colony1Name, "test_user1", env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.colony1Name, "test_user1", user1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetUser(env.colony2Name, "test_user2", env.executor2PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveUserSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	user1, user1PrvKey, err := utils.CreateTestUserWithKey(env.colony1Name, "test_user1")
	assert.Nil(t, err)
	_, err = client.AddUser(user1, env.colony1PrvKey)
	assert.Nil(t, err)

	user2, user2PrvKey, err := utils.CreateTestUserWithKey(env.colony2Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, env.colony2PrvKey)
	assert.Nil(t, err)

	err = client.RemoveUser(env.colony1Name, "test_user1", env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.colony1Name, "test_user1", env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.colony1Name, "test_user1", user1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.colony1Name, "test_user1", user2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.colony1Name, "test_user1", env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveUser(env.colony1Name, "test_user1", env.colony1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
