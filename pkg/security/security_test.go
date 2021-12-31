package security

import (
	"colonies/pkg/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverID(t *testing.T) {
	prvKey, err := GeneratePrivateKey()
	assert.Nil(t, err)

	id, err := GenerateID(prvKey)

	msg := "test_msg"
	signature, err := GenerateSignature(msg, prvKey)
	assert.Nil(t, err)

	recoveredID, err := RecoverID(msg, signature)
	assert.Nil(t, err)
	assert.Equal(t, recoveredID, id)
}

func TestRequireRoot(t *testing.T) {
	rootPassword := "password"
	assert.Nil(t, RequireRoot(rootPassword, rootPassword))
	assert.NotNil(t, RequireRoot(rootPassword, ""))
	assert.NotNil(t, RequireRoot(rootPassword, "invalid"))
}

func TestRequireColonyOwner(t *testing.T) {
	ownership := CreateOwnershipMock()
	colonyID := core.GenerateRandomID()
	ownership.addColony(colonyID)
	assert.Nil(t, RequireColonyOwner(colonyID, colonyID, ownership))
	assert.NotNil(t, RequireColonyOwner(core.GenerateRandomID(), colonyID, ownership))
}

func TestRequireRuntimeMembership(t *testing.T) {
	ownership := CreateOwnershipMock()
	colonyID := core.GenerateRandomID()
	ownership.addColony(colonyID)
	runtime1ID := core.GenerateRandomID()
	runtime2ID := core.GenerateRandomID()
	ownership.addRuntime(runtime1ID, colonyID)
	assert.NotNil(t, RequireRuntimeMembership(runtime1ID, colonyID, ownership)) // Should not work, not approved
	assert.NotNil(t, RequireRuntimeMembership(runtime2ID, colonyID, ownership)) // Should not work, not added or approved

	ownership.approveRuntime(runtime1ID, colonyID)
	ownership.approveRuntime(runtime2ID, colonyID)

	assert.Nil(t, RequireRuntimeMembership(runtime1ID, colonyID, ownership))    // Should work
	assert.NotNil(t, RequireRuntimeMembership(runtime2ID, colonyID, ownership)) // Should not work, not approved
}
