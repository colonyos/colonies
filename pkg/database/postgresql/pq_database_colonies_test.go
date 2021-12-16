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
	assert.Equal(t, colony.ID(), colonyFromDB.ID())
	assert.Equal(t, colony.Name(), colonyFromDB.Name())
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

	colonyFromDB, err := db.GetColonyByID(colony1.ID())
	assert.Nil(t, err)
	assert.Equal(t, colony1.ID(), colonyFromDB.ID())
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

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony1.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	assert.Nil(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony1.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	assert.Nil(t, err)

	worker3ID := core.GenerateRandomID()
	worker3 := core.CreateWorker(worker3ID, "test_worker", colony2.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker3)
	assert.Nil(t, err)

	err = db.DeleteColonyByID(colony1.ID())
	assert.Nil(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID())
	assert.Nil(t, err)
	assert.Nil(t, colonyFromDB)

	workerFromDB, err := db.GetWorkerByID(worker1.ID())
	assert.Nil(t, err)
	assert.Nil(t, workerFromDB)

	workerFromDB, err = db.GetWorkerByID(worker2.ID())
	assert.Nil(t, err)
	assert.Nil(t, workerFromDB)

	workerFromDB, err = db.GetWorkerByID(worker3.ID())
	assert.Nil(t, err)
	assert.NotNil(t, workerFromDB) // Belongs to colony 2 and should therefore not be deleted
}
