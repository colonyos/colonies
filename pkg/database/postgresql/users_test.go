package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddUser(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyName := core.GenerateRandomID()

	err = db.AddUser(nil)
	assert.NotNil(t, err) // Error

	user := utils.CreateTestUser(colonyName, "user1")
	err = db.AddUser(user)
	assert.Nil(t, err)

	err = db.AddUser(user) // Try to add an already existing user
	assert.NotNil(t, err)  // Error

	user2 := utils.CreateTestUser(colonyName, "user1") // username not unique
	err = db.AddUser(user2)
	assert.NotNil(t, err) // Error

	user2.Name = "unique_name"
	err = db.AddUser(user2)
	assert.Nil(t, err)

	userFromDB, err := db.GetUserByID(colonyName, user.ID)
	assert.Nil(t, err)
	assert.True(t, userFromDB.Equals(user))

	userFromDB, err = db.GetUserByName(colonyName, user.Name)
	assert.Nil(t, err)
	assert.True(t, userFromDB.Equals(user))

	defer db.Close()
}

func TestGetUsers(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyName := core.GenerateRandomID()

	user := utils.CreateTestUser(colonyName, "user1")
	err = db.AddUser(user)
	assert.Nil(t, err)

	user = utils.CreateTestUser(colonyName, "user2")
	err = db.AddUser(user)
	assert.Nil(t, err)

	users, err := db.GetUsersByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, users, 2)

	defer db.Close()
}

func TestRemoveUser(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyName := core.GenerateRandomID()

	user1 := utils.CreateTestUser(colonyName, "user1")
	err = db.AddUser(user1)
	assert.Nil(t, err)

	user2 := utils.CreateTestUser(colonyName, "user2")
	err = db.AddUser(user2)
	assert.Nil(t, err)

	user3 := utils.CreateTestUser(colonyName, "user3")
	err = db.AddUser(user3)
	assert.Nil(t, err)

	users, err := db.GetUsersByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, users, 3)

	err = db.RemoveUserByName(colonyName, user1.Name)
	assert.Nil(t, err)

	users, err = db.GetUsersByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, users, 2)

	user1FromDB, err := db.GetUserByName(colonyName, user1.Name)
	assert.Nil(t, err)
	assert.Nil(t, user1FromDB)

	user2FromDB, err := db.GetUserByName(colonyName, user2.Name)
	assert.Nil(t, err)
	assert.NotNil(t, user2FromDB)

	err = db.RemoveUserByName(colonyName, "user2")
	assert.Nil(t, err)

	user2FromDB, err = db.GetUserByName(colonyName, user2.Name)
	assert.Nil(t, err)
	assert.Nil(t, user2FromDB)

	users, err = db.GetUsersByColonyName(colonyName)
	assert.Nil(t, err)
	assert.Len(t, users, 1)

	user3FromDB, err := db.GetUserByName(colonyName, "user3")
	assert.Nil(t, err)
	assert.NotNil(t, user3FromDB)

	defer db.Close()
}

func TestRemoveUsersByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	user1 := utils.CreateTestUser(colonyName1, "user1")
	err = db.AddUser(user1)
	assert.Nil(t, err)

	user2 := utils.CreateTestUser(colonyName1, "user2")
	err = db.AddUser(user2)
	assert.Nil(t, err)

	user3 := utils.CreateTestUser(colonyName2, "user3")
	err = db.AddUser(user3)
	assert.Nil(t, err)

	users, err := db.GetUsersByColonyName(colonyName1)
	assert.Nil(t, err)
	assert.Len(t, users, 2)

	err = db.RemoveUsersByColonyName(colonyName1)
	assert.Nil(t, err)

	users, err = db.GetUsersByColonyName(colonyName1)
	assert.Nil(t, err)
	assert.Len(t, users, 0)

	users, err = db.GetUsersByColonyName(colonyName2)
	assert.Nil(t, err)
	assert.Len(t, users, 1)

	defer db.Close()
}
