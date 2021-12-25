package postgresql

import (
	"colonies/pkg/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddColony(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	colonies, err := db.GetColonies()
	assert.Nil(t, err)

	colonyFromDB := colonies[0]
	assert.True(t, colony.Equals(colonyFromDB))

	colonyFromDB, err = db.GetColonyByID(colony.ID)
	assert.Nil(t, err)
	assert.True(t, colony.Equals(colonyFromDB))
}

func TestAddTwoColonies(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	err = db.AddColony(colony2)
	assert.Nil(t, err)

	var colonies []*core.Colony
	colonies = append(colonies, colony1)
	colonies = append(colonies, colony2)

	coloniesFromDB, err := db.GetColonies()
	assert.Nil(t, err)
	assert.True(t, core.IsColonyArraysEqual(colonies, coloniesFromDB))
}

func TestGetColonyByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID)
	assert.Nil(t, err)
	assert.Equal(t, colony1.ID, colonyFromDB.ID)

	colonyFromDB, err = db.GetColonyByID(core.GenerateRandomID())
	assert.Nil(t, err)
}

func TestDeleteColonies(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	assert.Nil(t, err)

	runtime1ID := core.GenerateRandomID()
	runtime1 := core.CreateRuntime(runtime1ID, "test_runtime", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddRuntime(runtime1)
	assert.Nil(t, err)

	runtime2ID := core.GenerateRandomID()
	runtime2 := core.CreateRuntime(runtime2ID, "test_runtime", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddRuntime(runtime2)
	assert.Nil(t, err)

	runtime3ID := core.GenerateRandomID()
	runtime3 := core.CreateRuntime(runtime3ID, "test_runtime", colony2.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddRuntime(runtime3)
	assert.Nil(t, err)

	err = db.DeleteColonyByID(colony1.ID)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID)
	assert.Nil(t, err)
	assert.Nil(t, colonyFromDB)

	runtimeFromDB, err := db.GetRuntimeByID(runtime1.ID)
	assert.Nil(t, err)
	assert.Nil(t, runtimeFromDB)

	runtimeFromDB, err = db.GetRuntimeByID(runtime2.ID)
	assert.Nil(t, err)
	assert.Nil(t, runtimeFromDB)

	runtimeFromDB, err = db.GetRuntimeByID(runtime3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, runtimeFromDB) // Belongs to colony 2 and should therefore not be deleted
}
