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
	assert.Equal(t, colony.ID, colonyFromDB.ID)
	assert.Equal(t, colony.Name, colonyFromDB.Name)
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

	colonies, err := db.GetColonies()
	assert.Nil(t, err)
	assert.Len(t, colonies, 2)
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

	computer1ID := core.GenerateRandomID()
	computer1 := core.CreateComputer(computer1ID, "test_computer", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer1)
	assert.Nil(t, err)

	computer2ID := core.GenerateRandomID()
	computer2 := core.CreateComputer(computer2ID, "test_computer", colony1.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer2)
	assert.Nil(t, err)

	computer3ID := core.GenerateRandomID()
	computer3 := core.CreateComputer(computer3ID, "test_computer", colony2.ID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddComputer(computer3)
	assert.Nil(t, err)

	err = db.DeleteColonyByID(colony1.ID)
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID)
	assert.Nil(t, err)
	assert.Nil(t, colonyFromDB)

	computerFromDB, err := db.GetComputerByID(computer1.ID)
	assert.Nil(t, err)
	assert.Nil(t, computerFromDB)

	computerFromDB, err = db.GetComputerByID(computer2.ID)
	assert.Nil(t, err)
	assert.Nil(t, computerFromDB)

	computerFromDB, err = db.GetComputerByID(computer3.ID)
	assert.Nil(t, err)
	assert.NotNil(t, computerFromDB) // Belongs to colony 2 and should therefore not be deleted
}
