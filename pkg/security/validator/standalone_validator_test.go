package validator

import (
	"os"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/database/embedded"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/stretchr/testify/assert"
)

func prepareTestDB(prefix string) (database.Database, error) {
	if os.Getenv("COLONIES_DB_TYPE") == "embedded" {
		dir, err := os.MkdirTemp("", "colonies-validator-test-"+prefix+"-*")
		if err != nil {
			return nil, err
		}
		db := embedded.CreateEmbeddedDatabase(dir)
		if err := db.Initialize(); err != nil {
			return nil, err
		}
		return db, nil
	}
	if prefix == "" {
		prefix = "TEST_"
	}
	return postgresql.PrepareTestsWithPrefix(prefix)
}

func TestCreateStandaloneValidator(t *testing.T) {
	db, err := prepareTestDB("")
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
	ownership.addColony(colonyID, "my_colony")
	assert.Nil(t, security.RequireColonyOwner(colonyID, "my_colony"))
	assert.NotNil(t, security.RequireColonyOwner(core.GenerateRandomID(), colonyID))
}

func TestRequireMembership(t *testing.T) {
	ownership := createOwnershipMock()
	security := createTestValidator(ownership)

	colonyID := core.GenerateRandomID()
	ownership.addColony(colonyID, "my_colony")
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()
	ownership.addExecutor(executor1ID, colonyID)
	assert.NotNil(t, security.RequireMembership(executor1ID, colonyID, true)) // Should not work, not approved
	assert.Nil(t, security.RequireMembership(executor1ID, colonyID, false))   // Should work
	assert.NotNil(t, security.RequireMembership(executor2ID, colonyID, true)) // Should not work, not added or approved

	ownership.addExecutor(executor2ID, colonyID)
	ownership.approveExecutor(executor1ID, colonyID)

	assert.Nil(t, security.RequireMembership(executor1ID, colonyID, true))    // Should work
	assert.NotNil(t, security.RequireMembership(executor2ID, colonyID, true)) // Should not work, not approved
}
