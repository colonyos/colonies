package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestUserClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	user := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "test_user",
		ColonyName: "test_colony",
		Email:      "test@example.com",
	}

	// KVStore operations work even after close (in-memory store)
	err = db.AddUser(user)
	assert.Nil(t, err)

	_, err = db.GetUsers()
	assert.Nil(t, err)

	_, err = db.GetUserByID("test_colony", "invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing

	_, err = db.GetUserByName("test_colony", "invalid_name")
	assert.NotNil(t, err) // Expected error for non-existing

	_, err = db.GetUsersByColonyName("test_colony")
	assert.Nil(t, err) // Returns users or empty slice

	err = db.RemoveUserByID("test_colony", "invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing

	err = db.RemoveUserByName("test_colony", "invalid_name")
	assert.NotNil(t, err) // Expected error for non-existing

	err = db.RemoveUsersByColonyName("test_colony")
	assert.Nil(t, err) // No error when nothing to remove

	_, err = db.CountUsers()
	assert.Nil(t, err)

	_, err = db.CountUsersByColonyName("test_colony")
	assert.Nil(t, err)
}

func TestAddUser(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	user := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "test_user",
		ColonyName: colony.Name,
		Email:      "test@example.com",
	}

	// Test adding nil user
	err = db.AddUser(nil)
	assert.NotNil(t, err)

	// Test adding valid user
	err = db.AddUser(user)
	assert.Nil(t, err)

	// Test duplicate user
	err = db.AddUser(user)
	assert.NotNil(t, err)

	// Verify user was added
	userFromDB, err := db.GetUserByID(colony.Name, user.ID)
	assert.Nil(t, err)
	assert.True(t, user.Equals(userFromDB))

	// Test GetUserByName
	userFromDB, err = db.GetUserByName(colony.Name, user.Name)
	assert.Nil(t, err)
	assert.True(t, user.Equals(userFromDB))

	// Test GetUsers
	users, err := db.GetUsers()
	assert.Nil(t, err)
	assert.Len(t, users, 1)
	assert.True(t, user.Equals(users[0]))
}

func TestGetUserByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	user := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "test_user",
		ColonyName: colony.Name,
		Email:      "test@example.com",
	}

	err = db.AddUser(user)
	assert.Nil(t, err)

	userFromDB, err := db.GetUserByID(colony.Name, user.ID)
	assert.Nil(t, err)
	assert.True(t, user.Equals(userFromDB))

	// Test non-existing user
	_, err = db.GetUserByID(colony.Name, "non_existing_id")
	assert.NotNil(t, err)

	// Test invalid colony
	_, err = db.GetUserByID("invalid_colony", user.ID)
	assert.NotNil(t, err)
}

func TestGetUserByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	user := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "test_user",
		ColonyName: colony.Name,
		Email:      "test@example.com",
	}

	err = db.AddUser(user)
	assert.Nil(t, err)

	userFromDB, err := db.GetUserByName(colony.Name, user.Name)
	assert.Nil(t, err)
	assert.True(t, user.Equals(userFromDB))

	// Test non-existing user
	_, err = db.GetUserByName(colony.Name, "non_existing_name")
	assert.NotNil(t, err)

	// Test invalid colony
	_, err = db.GetUserByName("invalid_colony", user.Name)
	assert.NotNil(t, err)
}

func TestGetUsersByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "colony1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "colony2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	// Add users to different colonies
	user1 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user1",
		ColonyName: colony1.Name,
		Email:      "user1@example.com",
	}
	err = db.AddUser(user1)
	assert.Nil(t, err)

	user2 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user2",
		ColonyName: colony1.Name,
		Email:      "user2@example.com",
	}
	err = db.AddUser(user2)
	assert.Nil(t, err)

	user3 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user3",
		ColonyName: colony2.Name,
		Email:      "user3@example.com",
	}
	err = db.AddUser(user3)
	assert.Nil(t, err)

	// Test get users by colony
	users1, err := db.GetUsersByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.Len(t, users1, 2)

	users2, err := db.GetUsersByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.Len(t, users2, 1)

	// Test invalid colony
	invalidUsers, err := db.GetUsersByColonyName("invalid_colony")
	assert.Nil(t, err)
	assert.Empty(t, invalidUsers)
}

func TestRemoveUser(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	user := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "test_user",
		ColonyName: colony.Name,
		Email:      "test@example.com",
	}

	err = db.AddUser(user)
	assert.Nil(t, err)

	// Verify user exists
	count, err := db.CountUsersByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	// Remove user by ID
	err = db.RemoveUserByID(colony.Name, user.ID)
	assert.Nil(t, err)

	// Verify user is gone
	count, err = db.CountUsersByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	_, err = db.GetUserByID(colony.Name, user.ID)
	assert.NotNil(t, err)

	// Test remove non-existing user
	err = db.RemoveUserByID(colony.Name, "non_existing_id")
	assert.NotNil(t, err)

	// Test remove by name
	user2 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "test_user2",
		ColonyName: colony.Name,
		Email:      "test2@example.com",
	}
	err = db.AddUser(user2)
	assert.Nil(t, err)

	err = db.RemoveUserByName(colony.Name, user2.Name)
	assert.Nil(t, err)

	_, err = db.GetUserByName(colony.Name, user2.Name)
	assert.NotNil(t, err)

	// Test remove from invalid colony
	err = db.RemoveUserByID("invalid_colony", user.ID)
	assert.NotNil(t, err)
}

func TestRemoveUsersByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Add multiple users
	user1 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user1",
		ColonyName: colony.Name,
		Email:      "user1@example.com",
	}
	err = db.AddUser(user1)
	assert.Nil(t, err)

	user2 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user2",
		ColonyName: colony.Name,
		Email:      "user2@example.com",
	}
	err = db.AddUser(user2)
	assert.Nil(t, err)

	// Verify users exist
	count, err := db.CountUsersByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 2)

	// Remove all users
	err = db.RemoveUsersByColonyName(colony.Name)
	assert.Nil(t, err)

	// Verify all users are gone
	count, err = db.CountUsersByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	users, err := db.GetUsersByColonyName(colony.Name)
	assert.Nil(t, err)
	assert.Empty(t, users)

	// Test remove from invalid colony - should not error
	err = db.RemoveUsersByColonyName("invalid_colony")
	assert.Nil(t, err)
}

func TestCountUsers(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "colony1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "colony2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	// Add users to different colonies
	user1 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user1",
		ColonyName: colony1.Name,
		Email:      "user1@example.com",
	}
	err = db.AddUser(user1)
	assert.Nil(t, err)

	user2 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user2",
		ColonyName: colony1.Name,
		Email:      "user2@example.com",
	}
	err = db.AddUser(user2)
	assert.Nil(t, err)

	user3 := &core.User{
		ID:         core.GenerateRandomID(),
		Name:       "user3",
		ColonyName: colony2.Name,
		Email:      "user3@example.com",
	}
	err = db.AddUser(user3)
	assert.Nil(t, err)

	// Test total count
	totalCount, err := db.CountUsers()
	assert.Nil(t, err)
	assert.Equal(t, totalCount, 3)

	// Test colony-specific counts
	count1, err := db.CountUsersByColonyName(colony1.Name)
	assert.Nil(t, err)
	assert.Equal(t, count1, 2)

	count2, err := db.CountUsersByColonyName(colony2.Name)
	assert.Nil(t, err)
	assert.Equal(t, count2, 1)

	// Test invalid colony
	invalidCount, err := db.CountUsersByColonyName("invalid_colony")
	assert.Nil(t, err)
	assert.Equal(t, invalidCount, 0)
}