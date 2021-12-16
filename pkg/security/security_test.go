package security

import (
	"colonies/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyAPIKey(t *testing.T) {
	apiKey := "apikey"
	assert.Nil(t, VerifyAPIKey(apiKey, apiKey))
	assert.NotNil(t, VerifyAPIKey(apiKey, ""))
	assert.NotNil(t, VerifyAPIKey(apiKey, "invalid"))
}

func TestVerifyColonyOwnership(t *testing.T) {
	idendity, err := crypto.CreateIdendity()
	message := "test_message"
	colonyID := idendity.ID()
	assert.Nil(t, err)
	hash := crypto.GenerateHash([]byte(message))
	signature, err := crypto.Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	ownership := CreateOwnershipMock()
	err = VerifyColonyOwnership(colonyID, message, string(signature), ownership)
	assert.NotNil(t, err) // Should be an error since colony does not exists

	ownership.AddColony(colonyID)
	err = VerifyColonyOwnership(colonyID, message, string(signature), ownership)
	assert.Nil(t, err) // Should work now

	// Use an invalid cert
	ownership.AddColony(colonyID)
	err = VerifyColonyOwnership(colonyID, message, "", ownership)
	assert.NotNil(t, err) // Whould not work

	idendity2, err := crypto.CreateIdendity()
	assert.Nil(t, err)
	signature2, err := crypto.Sign(hash, idendity2.PrivateKey())
	assert.Nil(t, err)

	ownership.AddColony(colonyID)
	err = VerifyColonyOwnership(colonyID, message, string(signature2), ownership)
	assert.NotNil(t, err) // Should not work
}
