package database

import (
	"colonies/pkg/core"
	. "colonies/pkg/utils"
	"testing"
)

func TestAddColony(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	CheckError(t, err)

	colonies, err := db.GetColonies()
	CheckError(t, err)

	colonyFromDB := colonies[0]
	if colonyFromDB.ID() != colony.ID() {
		Fatal(t, "colony id mismatches")
	}
	if colonyFromDB.Name() != colony.Name() {
		Fatal(t, "name mismatches")
	}
}

func TestAddTwoColonies(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	CheckError(t, err)

	colonies, err := db.GetColonies()
	CheckError(t, err)
	if len(colonies) != 2 {
		Fatal(t, "invalid size of colonies array, expected 2")
	}
}

func TestGetColonyByID(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	CheckError(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID())
	CheckError(t, err)
	if colonyFromDB.ID() != colony1.ID() {
		Fatal(t, "failed to get colony by id ")
	}
}

func TestDeleteColonies(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")

	err = db.AddColony(colony2)
	CheckError(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony1.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony1.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	CheckError(t, err)

	worker3ID := core.GenerateRandomID()
	worker3 := core.CreateWorker(worker3ID, "test_worker", colony2.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker3)
	CheckError(t, err)

	err = db.DeleteColonyByID(colony1.ID())
	CheckError(t, err)

	colonyFromDB, err := db.GetColonyByID(colony1.ID())
	CheckError(t, err)
	if colonyFromDB != nil {
		Fatal(t, "expected colony to be nil")
	}

	workerFromDB, err := db.GetWorkerByID(worker1.ID())
	CheckError(t, err)
	if workerFromDB != nil {
		Fatal(t, "expected worker to be nil")
	}

	workerFromDB, err = db.GetWorkerByID(worker2.ID())
	CheckError(t, err)
	if workerFromDB != nil {
		Fatal(t, "expected worker to be nil")
	}

	workerFromDB, err = db.GetWorkerByID(worker3.ID())
	CheckError(t, err)
	if workerFromDB == nil {
		Fatal(t, "expected worker not to be nil")
	}
}
