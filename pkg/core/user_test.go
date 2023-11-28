package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	colonyName := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	email := "test@test.com"
	phone := "12345677"
	user := CreateUser(colonyName, userID, name, email, phone)

	assert.Equal(t, user.Name, name)
	assert.Len(t, user.ID, 64)
}

func TestUserToJSON(t *testing.T) {
	colonyName := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	email := "test@test.com"
	phone := "12345677"
	user := CreateUser(colonyName, userID, name, email, phone)

	jsonString, err := user.ToJSON()
	assert.Nil(t, err)

	_, err = ConvertJSONToUser(jsonString + "error")
	assert.NotNil(t, err)

	user2, err := ConvertJSONToUser(jsonString)
	assert.Nil(t, err)
	assert.True(t, user2.Equals(user))
}

func TestUserEquals(t *testing.T) {
	colonyName := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	email := "test@test.com"
	phone := "12345677"
	user1 := CreateUser(colonyName, userID, name, email, phone)

	user2 := CreateUser(colonyName+"X", userID, name, email, phone)
	assert.False(t, user2.Equals(user1))
	user3 := CreateUser(colonyName, userID+"X", name, email, phone)
	assert.False(t, user3.Equals(user1))
	user4 := CreateUser(colonyName, userID, name+"X", email, phone)
	assert.False(t, user4.Equals(user1))
	assert.True(t, user1.Equals(user1))
}

func TestIsUserArraysEqual(t *testing.T) {
	colonyName := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	email := "test@test.com"
	phone := "12345677"
	user1 := CreateUser(colonyName, userID, name+"1", email, phone)
	user2 := CreateUser(colonyName, userID, name+"2", email, phone)

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

	colonyName := GenerateRandomID()
	userID := GenerateRandomID()
	name := "test_user_name"
	email := "test@test.com"
	phone := "12345677"
	user1 := CreateUser(colonyName, userID, name+"1", email, phone)
	user2 := CreateUser(colonyName, userID, name+"2", email, phone)

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
