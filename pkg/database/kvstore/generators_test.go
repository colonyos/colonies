package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGeneratorClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	generator := utils.FakeGenerator(t, core.GenerateRandomID(), "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	
	// KVStore operations work even after close (in-memory store)
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	err = db.SetGeneratorLastRun("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.SetGeneratorFirstPack("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	_, err = db.GetGeneratorByID("invalid_id")
	assert.Nil(t, err) // Returns nil for non-existing

	_, err = db.GetGeneratorByName("invalid_colony_name", "invalid_name")
	assert.Nil(t, err) // Returns nil for non-existing

	_, err = db.FindGeneratorsByColonyName("invalid_name", 100)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindAllGenerators()
	assert.Nil(t, err) // Returns generators

	err = db.RemoveGeneratorByID("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveAllGeneratorsByColonyName("invalid_name")
	assert.Nil(t, err) // No error when nothing to remove
}

func TestAddGenerator(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	generator := utils.FakeGenerator(t, core.GenerateRandomID(), "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	// Verify generator was added
	generatorFromDB, err := db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)
	assert.True(t, generator.Equals(generatorFromDB))

	// Test adding generator with same ID should work (update)
	generator.Name = "updated_name"
	err = db.AddGenerator(generator)
	assert.Nil(t, err)
}

func TestGetGeneratorByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	generator := utils.FakeGenerator(t, core.GenerateRandomID(), "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	// Test non-existing ID
	generatorFromDB, err := db.GetGeneratorByID("invalid_id")
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	// Test existing ID
	generatorFromDB, err = db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.True(t, generator.Equals(generatorFromDB))
}

func TestGetGeneratorByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	generator := utils.FakeGenerator(t, core.GenerateRandomID(), "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	generator.Name = "test_name"
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	// Test invalid name
	generatorFromDB, err := db.GetGeneratorByName(generator.ColonyName, "invalid_name")
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	// Test invalid colony
	generatorFromDB, err = db.GetGeneratorByName("invalid_colony", "test_name")
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	// Test valid name and colony
	generatorFromDB, err = db.GetGeneratorByName(generator.ColonyName, "test_name")
	assert.Nil(t, err)
	assert.True(t, generator.Equals(generatorFromDB))
}

func TestSetGeneratorLastRun(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	generator := utils.FakeGenerator(t, core.GenerateRandomID(), "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.True(t, generator.Equals(generatorFromDB))

	lastRun := generatorFromDB.LastRun.Unix()

	err = db.SetGeneratorLastRun(generator.ID)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)

	assert.Greater(t, generatorFromDB.LastRun.Unix(), lastRun)

	// Test non-existing generator
	err = db.SetGeneratorLastRun("invalid_id")
	assert.NotNil(t, err)
}

func TestSetGeneratorFirstPack(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	generator := utils.FakeGenerator(t, core.GenerateRandomID(), "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.True(t, generator.Equals(generatorFromDB))

	err = db.SetGeneratorFirstPack(generator.ID)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)

	assert.True(t, generatorFromDB.FirstPack.Unix() > 0)

	// Test non-existing generator
	err = db.SetGeneratorFirstPack("invalid_id")
	assert.NotNil(t, err)
}

func TestFindGeneratorsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	generator1 := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator1.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator2.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	// Add generator from different colony
	otherColony := core.GenerateRandomID()
	generator3 := utils.FakeGenerator(t, otherColony, "test_initiator_id", "test_initiator_name")
	generator3.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator3)
	assert.Nil(t, err)

	generatorsFromDB, err := db.FindGeneratorsByColonyName(colonyName, 100)
	assert.Nil(t, err)
	assert.Len(t, generatorsFromDB, 2)

	count := 0
	for _, generator := range generatorsFromDB {
		if generator.ID == generator1.ID {
			count++
		}
		if generator.ID == generator2.ID {
			count++
		}
		assert.Equal(t, generator.ColonyName, colonyName)
	}
	assert.True(t, count == 2)

	// Test with limit
	generatorsLimited, err := db.FindGeneratorsByColonyName(colonyName, 1)
	assert.Nil(t, err)
	assert.Len(t, generatorsLimited, 1)

	// Test non-existing colony
	generatorsEmpty, err := db.FindGeneratorsByColonyName("non_existing_colony", 100)
	assert.Nil(t, err)
	assert.Empty(t, generatorsEmpty)
}

func TestFindAllGenerators(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	generator1 := utils.FakeGenerator(t, colonyName1, "test_initiator_id", "test_initiator_name")
	generator1.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	colonyName2 := core.GenerateRandomID()
	generator2 := utils.FakeGenerator(t, colonyName2, "test_initiator_id", "test_initiator_name")
	generator2.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	generatorsFromDB, err := db.FindAllGenerators()
	assert.Nil(t, err)
	assert.Len(t, generatorsFromDB, 2)

	// Verify both generators are returned
	found := 0
	for _, gen := range generatorsFromDB {
		if gen.ID == generator1.ID || gen.ID == generator2.ID {
			found++
		}
	}
	assert.Equal(t, found, 2)
}

func TestRemoveGeneratorByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	generator1 := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator1.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator2.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	// Add generator args to test cascade delete
	generatorArg := core.CreateGeneratorArg(generator1.ID, colonyName, "arg")
	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generator1.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	err = db.RemoveGeneratorByID(generator1.ID)
	assert.Nil(t, err)

	generatorFromDB, err = db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	// Verify other generator still exists
	generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	// Verify generator args were also removed
	count, err = db.CountGeneratorArgs(generator1.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	// Test removing non-existing generator
	err = db.RemoveGeneratorByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestRemoveAllGeneratorsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	generator1 := utils.FakeGenerator(t, colonyName1, "test_initiator_id", "test_initiator_name")
	generator1.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator1)
	assert.Nil(t, err)

	generator2 := utils.FakeGenerator(t, colonyName1, "test_initiator_id", "test_initiator_name")
	generator2.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	colonyName2 := core.GenerateRandomID()
	generator3 := utils.FakeGenerator(t, colonyName2, "test_initiator_id", "test_initiator_name")
	generator3.ID = core.GenerateRandomID()
	err = db.AddGenerator(generator3)
	assert.Nil(t, err)

	// Add generator args for testing cascade delete
	generatorArg := core.CreateGeneratorArg(generator1.ID, colonyName1, "arg")
	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)
	generatorArg = core.CreateGeneratorArg(generator2.ID, colonyName1, "arg")
	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)
	generatorArg = core.CreateGeneratorArg(generator3.ID, colonyName2, "arg")
	err = db.AddGeneratorArg(generatorArg)
	assert.Nil(t, err)

	count, err := db.CountGeneratorArgs(generator1.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	generatorFromDB, err := db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	err = db.RemoveAllGeneratorsByColonyName(colonyName1)
	assert.Nil(t, err)

	// Verify generators from colony1 are removed
	generatorFromDB, err = db.GetGeneratorByID(generator1.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	generatorFromDB, err = db.GetGeneratorByID(generator2.ID)
	assert.Nil(t, err)
	assert.Nil(t, generatorFromDB)

	// Verify generator from colony2 still exists
	generatorFromDB, err = db.GetGeneratorByID(generator3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, generatorFromDB)

	// Verify generator args were also removed appropriately
	count, err = db.CountGeneratorArgs(generator1.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generator2.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountGeneratorArgs(generator3.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	// Test removing from non-existing colony - should not error
	err = db.RemoveAllGeneratorsByColonyName("non_existing_colony")
	assert.Nil(t, err)
}

func TestGeneratorComplexScenarios(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test generator with complex trigger
	colonyName := core.GenerateRandomID()
	generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	generator.Name = "complex_generator"
	generator.Trigger = 3600 // 1 hour trigger
	
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	// Test updating generator fields
	err = db.SetGeneratorLastRun(generator.ID)
	assert.Nil(t, err)
	err = db.SetGeneratorFirstPack(generator.ID)
	assert.Nil(t, err)

	// Verify updates
	updatedGen, err := db.GetGeneratorByID(generator.ID)
	assert.Nil(t, err)
	assert.NotNil(t, updatedGen)
	assert.Greater(t, updatedGen.LastRun.Unix(), generator.LastRun.Unix())
	assert.Greater(t, updatedGen.FirstPack.Unix(), generator.FirstPack.Unix())

	// Test multiple generators with same name in different colonies
	otherColony := core.GenerateRandomID()
	generator2 := utils.FakeGenerator(t, otherColony, "test_initiator_id", "test_initiator_name")
	generator2.ID = core.GenerateRandomID()
	generator2.Name = "complex_generator" // Same name, different colony
	
	err = db.AddGenerator(generator2)
	assert.Nil(t, err)

	// Both should be retrievable by their colony+name combination
	gen1, err := db.GetGeneratorByName(colonyName, "complex_generator")
	assert.Nil(t, err)
	assert.Equal(t, gen1.ID, generator.ID)

	gen2, err := db.GetGeneratorByName(otherColony, "complex_generator")
	assert.Nil(t, err)
	assert.Equal(t, gen2.ID, generator2.ID)
}

func TestGeneratorCascadeOperations(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	generator := utils.FakeGenerator(t, colonyName, "test_initiator_id", "test_initiator_name")
	generator.ID = core.GenerateRandomID()
	
	err = db.AddGenerator(generator)
	assert.Nil(t, err)

	// Add multiple generator args
	arg1 := core.CreateGeneratorArg(generator.ID, colonyName, "arg1")
	arg2 := core.CreateGeneratorArg(generator.ID, colonyName, "arg2")
	arg3 := core.CreateGeneratorArg(generator.ID, colonyName, "arg3")

	err = db.AddGeneratorArg(arg1)
	assert.Nil(t, err)
	err = db.AddGeneratorArg(arg2)
	assert.Nil(t, err)
	err = db.AddGeneratorArg(arg3)
	assert.Nil(t, err)

	// Verify args exist
	count, err := db.CountGeneratorArgs(generator.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 3)

	// Remove generator should cascade delete all args
	err = db.RemoveGeneratorByID(generator.ID)
	assert.Nil(t, err)

	// Verify all args are removed
	count, err = db.CountGeneratorArgs(generator.ID)
	assert.Nil(t, err)
	assert.Equal(t, count, 0)
}