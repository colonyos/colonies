package validator

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/stretchr/testify/assert"
)

func TestCreateStandaloneValidator(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	validator := CreateValidator(db)
	assert.NotNil(t, validator)

	db.Close()
}

func TestRequireRoot(t *testing.T) {
	security := createTestValidator(createOwnershipMock())

	serverID := core.GenerateRandomID()
	assert.Nil(t, security.RequireServerOwner(serverID, serverID))
	assert.NotNil(t, security.RequireServerOwner(serverID, ""))
	assert.NotNil(t, security.RequireServerOwner(serverID, core.GenerateRandomID()))
}

func TestRequireColonyOwner(t *testing.T) {
	ownership := createOwnershipMock()
	security := createTestValidator(ownership)

	colonyID := core.GenerateRandomID()
	ownership.addColony(colonyID)
	assert.Nil(t, security.RequireColonyOwner(colonyID, colonyID))
	assert.NotNil(t, security.RequireColonyOwner(core.GenerateRandomID(), colonyID))
}

func TestRequireRuntimeMembership(t *testing.T) {
	ownership := createOwnershipMock()
	security := createTestValidator(ownership)

	colonyID := core.GenerateRandomID()
	ownership.addColony(colonyID)
	runtime1ID := core.GenerateRandomID()
	runtime2ID := core.GenerateRandomID()
	ownership.addRuntime(runtime1ID, colonyID)
	assert.NotNil(t, security.RequireRuntimeMembership(runtime1ID, colonyID, true)) // Should not work, not approved
	assert.Nil(t, security.RequireRuntimeMembership(runtime1ID, colonyID, false))   // Should work
	assert.NotNil(t, security.RequireRuntimeMembership(runtime2ID, colonyID, true)) // Should not work, not added or approved

	ownership.addRuntime(runtime2ID, colonyID)
	ownership.approveRuntime(runtime1ID, colonyID)

	assert.Nil(t, security.RequireRuntimeMembership(runtime1ID, colonyID, true))    // Should work
	assert.NotNil(t, security.RequireRuntimeMembership(runtime2ID, colonyID, true)) // Should not work, not approved
}
