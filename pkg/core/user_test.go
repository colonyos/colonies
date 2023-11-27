package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	colonyID := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	user := CreateUser(colonyID, userID, name)

	assert.Equal(t, user.Name, name)
	assert.Len(t, user.ID, 64)
}

func TestUserToJSON(t *testing.T) {
	colonyID := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	user := CreateUser(colonyID, userID, name)

	jsonString, err := user.ToJSON()
	assert.Nil(t, err)

	_, err = ConvertJSONToUser(jsonString + "error")
	assert.NotNil(t, err)

	user2, err := ConvertJSONToUser(jsonString)
	assert.Nil(t, err)
	assert.True(t, user2.Equals(user))
}

func TestUserEquals(t *testing.T) {
	colonyID := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	user1 := CreateUser(colonyID, userID, name)

	user2 := CreateUser(colonyID+"X", userID, name)
	assert.False(t, user2.Equals(user1))
	user3 := CreateUser(colonyID, userID+"X", name)
	assert.False(t, user3.Equals(user1))
	user4 := CreateUser(colonyID, userID, name+"X")
	assert.False(t, user4.Equals(user1))
	assert.True(t, user1.Equals(user1))
}

func TestIsUserArraysEqual(t *testing.T) {
	colonyID := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	user1 := CreateUser(colonyID, userID, name+"1")
	user2 := CreateUser(colonyID, userID, name+"2")

	var users1 []*User
	users1 = append(users1, user1)
	users1 = append(users1, user2)

	var users2 []*User
	users1 = append(users2, user1)

	assert.True(t, IsUserArraysEqual(users1, users1))
	assert.False(t, IsUserArraysEqual(users1, users2))
}

func TestUserToJSONArray(t *testing.T) {
	var users []*User

	colonyID := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	user1 := CreateUser(colonyID, userID, name+"1")
	user2 := CreateUser(colonyID, userID, name+"2")

	users = append(users, user1)
	users = append(users, user2)

	jsonString, err := ConvertUserArrayToJSON(users)
	assert.Nil(t, err)

	users2, err := ConvertJSONToUserArray(jsonString + "error")
	assert.NotNil(t, err)

	users2, err = ConvertJSONToUserArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsUserArraysEqual(users, users2))
}
