package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestColonyClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	// KVStore operations work even after close (in-memory store)
	err = db.AddColony(colony)
	assert.Nil(t, err)

	_, err = db.GetColonies()
	assert.Nil(t, err)

	_, err = db.GetColonyByID("invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing ID

	_, err = db.GetColonyByName("invalid_name")
	assert.Nil(t, err) // Returns nil, nil for non-existing

	err = db.RenameColony("invalid_id", "invalid_name")
	assert.NotNil(t, err) // Expected error for non-existing colony

	err = db.RemoveColonyByName("invalid_id")
	assert.NotNil(t, err) // Expected error for non-existing colony

	_, err = db.CountColonies()
	assert.Nil(t, err)
}

func TestAddColony(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	// Test adding nil colony
	err = db.AddColony(nil)
	assert.NotNil(t, err)

	// Test adding valid colony
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Test adding same colony again - should fail
	err = db.AddColony(colony)
	assert.NotNil(t, err)

	// Verify colony was added
	colonies, err := db.GetColonies()
	assert.Nil(t, err)
	assert.Len(t, colonies, 1)

	colonyFromDB := colonies[0]
	assert.True(t, colony.Equals(colonyFromDB))

	// Test GetColonyByID
	colonyFromDB, err = db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromDB))

	// Test GetColonyByName
	colonyFromDB, err = db.GetColonyByName(colony.Name)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromDB))
}

func TestRenameColony(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.Equal(t, colonyFromDB.Name, "test_colony_name")

	// Test rename
	err = db.RenameColony(colony.Name, "test_colony_new_name")
	assert.Nil(t, err)

	colonyFromDB, err = db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.Equal(t, colonyFromDB.Name, "test_colony_new_name")

	// Test rename non-existing colony
	err = db.RenameColony("non_existing", "new_name")
	assert.NotNil(t, err)

	// Test rename to existing name
	colony2 := core.CreateColony(core.GenerateRandomID(), "another_colony")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	err = db.RenameColony("another_colony", "test_colony_new_name")
	assert.NotNil(t, err)
}

func TestAddTwoColonies(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	colonies, err := db.GetColonies()
	assert.Nil(t, err)
	assert.Len(t, colonies, 2)

	count, err := db.CountColonies()
	assert.Nil(t, err)
	assert.Equal(t, count, 2)
}

func TestRemoveColony(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Verify colony exists
	count, err := db.CountColonies()
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	// Remove colony
	err = db.RemoveColonyByName(colony.Name)
	assert.Nil(t, err)

	// Verify colony is gone
	count, err = db.CountColonies()
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	// Try to get removed colony
	colonyFromDB, err := db.GetColonyByName(colony.Name)
	assert.Nil(t, err)
	assert.Nil(t, colonyFromDB)

	// Try to remove non-existing colony
	err = db.RemoveColonyByName("non_existing")
	assert.NotNil(t, err)
}

func TestGetColonyByIDNotFound(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Try to get non-existing colony by ID
	colony, err := db.GetColonyByID("non_existing_id")
	assert.NotNil(t, err)
	assert.Nil(t, colony)
}

func TestGetColonyByNameNotFound(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Try to get non-existing colony by name - should return (nil, nil)
	colony, err := db.GetColonyByName("non_existing_name")
	assert.Nil(t, err)
	assert.Nil(t, colony)
}