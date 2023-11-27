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

	colonyID := core.GenerateRandomID()

	user := utils.CreateTestUser(colonyID, "user1")
	err = db.AddUser(user)
	assert.Nil(t, err)

	userFromDB, err := db.GetUserByID(colonyID, user.ID)
	assert.Nil(t, err)
	assert.True(t, userFromDB.Equals(user))

	userFromDB, err = db.GetUserByName(colonyID, user.Name)
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

func TestDeleteUser(t *testing.T) {
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

	err = db.DeleteUserByName(colonyName, user1.Name)
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

	err = db.DeleteUserByName(colonyName, "user2")
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

func TestDeleteUsersByColonyName(t *testing.T) {
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

	err = db.DeleteUsersByColonyName(colonyName1)
	assert.Nil(t, err)

	users, err = db.GetUsersByColonyName(colonyName1)
	assert.Nil(t, err)
	assert.Len(t, users, 0)

	users, err = db.GetUsersByColonyName(colonyName2)
	assert.Nil(t, err)
	assert.Len(t, users, 1)

	defer db.Close()
}
