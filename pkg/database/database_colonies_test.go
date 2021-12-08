package database

import (
	"colonies/pkg/core"
	. "colonies/pkg/utils"
	"testing"
)

func TestAddColony(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name")
	CheckError(t, err)

	err = db.AddColony(colony)
	CheckError(t, err)

	colonies, err := db.GetColonies()
	CheckError(t, err)

	colonyFromDB := colonies[0]
	if colonyFromDB.ID() != colony.ID() {
		Fatal(t, "colony id mismatches")
	}
	if colonyFromDB.PrivateKey() != colony.PrivateKey() {
		Fatal(t, "private key mismatches")
	}
	if colonyFromDB.Name() != colony.Name() {
		Fatal(t, "name mismatches")
	}
}

func TestAddTwoColonies(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony1, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2, err := core.CreateColony("test_colony_name_2")
	CheckError(t, err)

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

	colony1, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2, err := core.CreateColony("test_colony_name_2")
	CheckError(t, err)

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

	colony1, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2, err := core.CreateColony("test_colony_name_2")
	CheckError(t, err)

	err = db.AddColony(colony2)
	CheckError(t, err)

	worker1 := core.CreateWorker("5dfda4f1d4be06bf9d9a143737fc87698e65f09c404c05de80ed43a49fe9aea", "test_worker", colony1.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2 := core.CreateWorker("4c9e02e0e1ee3e895128de093039d89cdeb7e66742520c96c4772afb374366a", "test_worker", colony1.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	CheckError(t, err)

	worker3 := core.CreateWorker("c93a222feea1e8d567a2c9d0f9e84bd7b3fe808bc5fd2e329cca0923184c096e", "test_worker", colony2.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

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
