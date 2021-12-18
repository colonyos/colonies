package security

import (
	"colonies/pkg/core"
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
	id := colonyID

	signature, err := GenerateSignature(message, idendity.PrivateKeyAsHex())
	assert.Nil(t, err)

	ownership := CreateOwnershipMock()
	err = VerifyColonyOwnership(id, colonyID, message, string(signature), ownership)
	assert.NotNil(t, err) // Should be an error since colony does not exists

	ownership.addColony(colonyID)
	err = VerifyColonyOwnership(id, colonyID, message, string(signature), ownership)
	assert.Nil(t, err) // Should work now

	// Use an invalid cert
	ownership.addColony(colonyID)
	err = VerifyColonyOwnership(id, colonyID, message, "", ownership)
	assert.NotNil(t, err) // Whould not work

	idendity2, err := crypto.CreateIdendity()
	assert.Nil(t, err)
	signature2, err := GenerateSignature(message, idendity2.PrivateKeyAsHex())
	assert.Nil(t, err)

	ownership.addColony(colonyID)
	err = VerifyColonyOwnership(id, colonyID, message, string(signature2), ownership)
	assert.NotNil(t, err) // Should not work
}

func TestVerifyWorkerMembership(t *testing.T) {
	colonyID := core.GenerateRandomID()
	ownership := CreateOwnershipMock()

	prvKey, err := GeneratePrivateKey()
	workerID, err := GenerateID(prvKey)
	assert.Nil(t, err)
	id := workerID
	assert.Nil(t, err)

	digest := GenerateDigest()
	sig, err := GenerateSignature(digest, prvKey)
	assert.Nil(t, err)

	err = VerifyWorkerMembership(id, colonyID, digest, sig, ownership)
	assert.NotNil(t, err) // Should not work since worker not member of colony

	ownership.addWorker(workerID, colonyID)
	err = VerifyWorkerMembership(workerID, colonyID, digest, sig, ownership)
	assert.Nil(t, err)
}
