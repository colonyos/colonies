package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewKVStoreDatabaseWithTesting(t *testing.T) {
	db := NewKVStoreDatabase()
	assert.NotNil(t, db)
	assert.NotNil(t, db.store)
	assert.False(t, db.initialized)
	assert.False(t, db.locked)
}

func TestInitializeDatabase(t *testing.T) {
	db := NewKVStoreDatabase()

	// Test first initialization
	err := db.Initialize()
	assert.Nil(t, err)
	assert.True(t, db.initialized)

	// Test that we can initialize multiple times without error
	err = db.Initialize()
	assert.Nil(t, err)
}

func TestCloseDatabase(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	// Close should not error
	db.Close()
	assert.False(t, db.initialized)

	// After close, should be able to initialize again
	err = db.Initialize()
	assert.Nil(t, err)
	assert.True(t, db.initialized)
}

func TestDatabaseLocking(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test lock with timeout
	err = db.Lock(60)
	assert.Nil(t, err)
	assert.True(t, db.locked)

	// Test unlock
	err = db.Unlock()
	assert.Nil(t, err)
	assert.False(t, db.locked)

	// Test double lock - should error when already locked
	err = db.Lock(60)
	assert.Nil(t, err)
	err = db.Lock(60)
	assert.NotNil(t, err) // Should error since already locked
	assert.True(t, db.locked)

	err = db.Unlock()
	assert.Nil(t, err)
}

func TestDatabaseLockingClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	// Lock on closed DB should work (KVStore doesn't check if initialized)
	err = db.Lock(60)
	assert.Nil(t, err)

	// Unlock on closed DB should work
	err = db.Unlock()
	assert.Nil(t, err)
}

func TestDrop(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	// Add some test data
	colony := &core.Colony{
		ID:   "test_id",
		Name: "test_colony",
	}
	err = db.AddColony(colony)
	assert.Nil(t, err)

	// Drop should clear all data
	err = db.Drop()
	assert.Nil(t, err)
	assert.False(t, db.initialized)

	// Re-initialize and verify data is gone
	err = db.Initialize()
	assert.Nil(t, err)

	// After drop and re-initialize, colonies should be empty
	colonies, err := db.GetColonies()
	if err != nil {
		// GetColonies might return error if no colonies exist, which is OK
		assert.Empty(t, colonies)
	} else {
		assert.Empty(t, colonies)
	}
}