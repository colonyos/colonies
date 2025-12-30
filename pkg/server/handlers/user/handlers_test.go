package user_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	addedColony, err := client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(addedColony))

	user := utils.CreateTestUser(colony.Name, "test_user")
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	s.Shutdown()
	<-done
}

func TestGetUsers(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

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

	s.Shutdown()
	<-done
}

func TestGetUser(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

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

	s.Shutdown()
	<-done
}

func TestRemoveUser(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

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

	err = client.RemoveUser(colony.Name, "test_user2", colonyPrvKey)
	assert.Nil(t, err)

	usersFromServer, err = client.GetUsers(colony.Name, prvKey)
	assert.Nil(t, err)
	assert.Len(t, usersFromServer, 1)

	s.Shutdown()
	<-done
}

// TestGetUserByID tests getting a user by their ID
func TestGetUserByID(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	// Get user by ID
	userFromServer, err := client.GetUserByID(colony.Name, addedUser.ID, prvKey)
	assert.Nil(t, err)
	assert.Equal(t, addedUser.ID, userFromServer.ID)
	assert.Equal(t, addedUser.Name, userFromServer.Name)

	s.Shutdown()
	<-done
}

// TestGetUserByIDNotFound tests getting a non-existent user by ID
func TestGetUserByIDNotFound(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to get non-existent user by ID
	_, err = client.GetUserByID(colony.Name, "nonexistent-user-id", prvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUserByIDUnauthorized tests that non-members cannot get user by ID
func TestGetUserByIDUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	// Add user to colony1
	user1, _, err := utils.CreateTestUserWithKey(colony1.Name, "test_user1")
	assert.Nil(t, err)
	addedUser, err := client.AddUser(user1, colonyPrvKey1)
	assert.Nil(t, err)

	// Add user to colony2
	user2, user2PrvKey, err := utils.CreateTestUserWithKey(colony2.Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to get user from colony1 using colony2 user's key
	_, err = client.GetUserByID(colony1.Name, addedUser.ID, user2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestAddUserDuplicate tests adding a user with duplicate name
func TestAddUserDuplicate(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user := utils.CreateTestUser(colony.Name, "test_user")
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to add same user again
	user2 := utils.CreateTestUser(colony.Name, "test_user")
	_, err = client.AddUser(user2, colonyPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestAddUserNotColonyOwner tests that non-owner cannot add user
func TestAddUserNotColonyOwner(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

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

	// Try to add user with executor key
	user := utils.CreateTestUser(colony.Name, "test_user")
	_, err = client.AddUser(user, executorPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestAddUserNonExistentColony tests adding user to non-existent colony
func TestAddUserNonExistentColony(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to add user to non-existent colony
	user := utils.CreateTestUser("nonexistent-colony", "test_user")
	_, err = client.AddUser(user, colonyPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUserNotFound tests getting a non-existent user
func TestGetUserNotFound(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to get non-existent user
	_, err = client.GetUser(colony.Name, "nonexistent-user", prvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUserUnauthorized tests that non-members cannot get user
func TestGetUserUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	// Add user to colony1
	user1 := utils.CreateTestUser(colony1.Name, "test_user1")
	_, err = client.AddUser(user1, colonyPrvKey1)
	assert.Nil(t, err)

	// Add user to colony2
	user2, user2PrvKey, err := utils.CreateTestUserWithKey(colony2.Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to get user from colony1 using colony2 user's key
	_, err = client.GetUser(colony1.Name, "test_user1", user2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUsersUnauthorized tests that non-members cannot get users list
func TestGetUsersUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, _, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	// Add user to colony2
	user2, user2PrvKey, err := utils.CreateTestUserWithKey(colony2.Name, "test_user2")
	assert.Nil(t, err)
	_, err = client.AddUser(user2, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to get users from colony1 using colony2 user's key
	_, err = client.GetUsers(colony1.Name, user2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUsersNonExistentColony tests getting users from non-existent colony
func TestGetUsersNonExistentColony(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to get users from non-existent colony
	_, err = client.GetUsers("nonexistent-colony", prvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveUserNotFound tests removing non-existent user
func TestRemoveUserNotFound(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to remove non-existent user
	err = client.RemoveUser(colony.Name, "nonexistent-user", colonyPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveUserNotColonyOwner tests that non-owner cannot remove user
func TestRemoveUserNotColonyOwner(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to remove user with user's own key (not colony owner)
	err = client.RemoveUser(colony.Name, "test_user", prvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveUserNonExistentColony tests removing user from non-existent colony
func TestRemoveUserNonExistentColony(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	// Try to remove user from non-existent colony
	err = client.RemoveUser("nonexistent-colony", "test_user", colonyPrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUsersEmpty tests getting users when none exist
func TestGetUsersEmpty(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

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

	users, err := client.GetUsers(colony.Name, executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, users, 0)

	s.Shutdown()
	<-done
}

// TestGetUserNonExistentColony tests getting user from non-existent colony
func TestGetUserNonExistentColony(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	_, err = client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to get user from non-existent colony
	_, err = client.GetUser("nonexistent-colony", "test_user", prvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestGetUserByIDNonExistentColony tests getting user by ID from non-existent colony
func TestGetUserByIDNonExistentColony(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	user, prvKey, err := utils.CreateTestUserWithKey(colony.Name, "test_user")
	assert.Nil(t, err)
	addedUser, err := client.AddUser(user, colonyPrvKey)
	assert.Nil(t, err)

	// Try to get user by ID from non-existent colony
	_, err = client.GetUserByID("nonexistent-colony", addedUser.ID, prvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}