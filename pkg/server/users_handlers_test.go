package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	user := utils.CreateTestUser(colony.Name, "test_user")
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	server.Shutdown()
	<-done
}

func TestGetUsers(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user1")
	assert.Nil(t, err)
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	user = utils.CreateTestUser(colony.Name, "test_user2")
	addedUser, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	usersFromServer, err := client.GetUsers(colony.Name, prvKey)
	assert.Nil(t, err)
	assert.Len(t, usersFromServer, 2)

	server.Shutdown()
	<-done
}

func TestGetUser(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user1")
	assert.Nil(t, err)
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	user = utils.CreateTestUser(colony.Name, "test_user2")
	addedUser, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	userFromServer, err := client.GetUser(colony.Name, "test_user1", prvKey)
	assert.Nil(t, err)
	assert.Equal(t, userFromServer.Name, "test_user1")

	server.Shutdown()
	<-done
}

func TestDeleteUser(t *testing.T) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user1")
	assert.Nil(t, err)
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	user = utils.CreateTestUser(colony.Name, "test_user2")
	addedUser, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	usersFromServer, err := client.GetUsers(colony.Name, prvKey)
	assert.Nil(t, err)
	assert.Len(t, usersFromServer, 2)

	err = client.DeleteUser(colony.Name, "test_user2", colonyPrvKey)
	assert.Nil(t, err)

	usersFromServer, err = client.GetUsers(colony.Name, prvKey)
	assert.Nil(t, err)
	assert.Len(t, usersFromServer, 1)

	server.Shutdown()
	<-done
}
