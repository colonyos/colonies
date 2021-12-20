package security

import (
	"colonies/pkg/core"
	"colonies/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequireRoot(t *testing.T) {
	rootPassword := "password"
	assert.Nil(t, RequireRoot(rootPassword, rootPassword))
	assert.NotNil(t, RequireRoot(rootPassword, ""))
	assert.NotNil(t, RequireRoot(rootPassword, "invalid"))
}

func TestRequireColonyOwner(t *testing.T) {
	idendity, err := crypto.CreateIdendity()
	message := "test_message"
	colonyID := idendity.ID()
	id := colonyID

	signature, err := GenerateSignature(message, idendity.PrivateKeyAsHex())
	assert.Nil(t, err)

	ownership := CreateOwnershipMock()
	err = RequireColonyOwner(id, colonyID, message, string(signature), ownership)
	assert.NotNil(t, err) // Should be an error since colony does not exists

	ownership.addColony(colonyID)
	err = RequireColonyOwner(id, colonyID, message, string(signature), ownership)
	assert.Nil(t, err) // Should work now

	// Use an invalid cert
	ownership.addColony(colonyID)
	err = RequireColonyOwner(id, colonyID, message, "", ownership)
	assert.NotNil(t, err) // Whould not work

	idendity2, err := crypto.CreateIdendity()
	assert.Nil(t, err)
	signature2, err := GenerateSignature(message, idendity2.PrivateKeyAsHex())
	assert.Nil(t, err)

	ownership.addColony(colonyID)
	err = RequireColonyOwner(id, colonyID, message, string(signature2), ownership)
	assert.NotNil(t, err) // Should not work
}

func TestRequireColonyMember(t *testing.T) {
	colonyID := core.GenerateRandomID()
	ownership := CreateOwnershipMock()

	prvKey, err := GeneratePrivateKey()
	computerID, err := GenerateID(prvKey)
	assert.Nil(t, err)
	id := computerID
	assert.Nil(t, err)

	digest := GenerateDigest()
	sig, err := GenerateSignature(digest, prvKey)
	assert.Nil(t, err)

	err = RequireColonyMember(id, colonyID, digest, sig, ownership)
	assert.NotNil(t, err) // Should not work since computer not member of colony

	ownership.addComputer(computerID, colonyID)
	err = RequireColonyMember(computerID, colonyID, digest, sig, ownership)
	assert.Nil(t, err)
}

func TestRequireColonyOwnerOrMember(t *testing.T) {
	colonyID := core.GenerateRandomID()
	ownership := CreateOwnershipMock()

	prvKey, err := GeneratePrivateKey()
	computerID, err := GenerateID(prvKey)
	assert.Nil(t, err)
	id := computerID
	assert.Nil(t, err)

	digest := GenerateDigest()
	sig, err := GenerateSignature(digest, prvKey)
	assert.Nil(t, err)

	err = RequireColonyOwnerOrMember(id, colonyID, digest, sig, ownership)
	assert.NotNil(t, err) // Should not work since computer not member of colony

	ownership.addComputer(computerID, colonyID)
	err = RequireColonyOwnerOrMember(computerID, colonyID, digest, sig, ownership)
	assert.Nil(t, err)
}
