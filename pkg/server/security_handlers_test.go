package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestChangeUserID(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	user, userPrvKey, err := utils.CreateTestUserWithKey(env.colonyName, "test_user")
	assert.Nil(t, err)

	addedUser, err := client.AddUser(user, env.colonyPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, addedUser)

	// Check if the user can use the API
	users, err := client.GetUsers(env.colonyName, userPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	// Change Id
	crypto := crypto.CreateCrypto()
	newUserPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	newUserID, err := crypto.GenerateID(newUserPrvKey)
	assert.Nil(t, err)

	err = client.ChangeUserID(env.colonyName, newUserID, userPrvKey)
	assert.Nil(t, err)

	// Check if the user can use the API with the new private key
	users, err = client.GetUsers(env.colonyName, newUserPrvKey)
	assert.Nil(t, err)
	assert.NotNil(t, users)

	// Check if the user can use the API with the OLD private key, should not work
	_, err = client.GetUsers(env.colonyName, userPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
