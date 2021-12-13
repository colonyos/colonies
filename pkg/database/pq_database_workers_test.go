package database

import (
	"colonies/pkg/core"
	. "colonies/pkg/utils"
	"testing"
)

func TestAddWorker(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony)
	CheckError(t, err)

	workerID := core.GenerateRandomID()
	worker := core.CreateWorker(workerID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker)
	CheckError(t, err)

	workers, err := db.GetWorkers()
	CheckError(t, err)

	workerFromDB := workers[0]

	if !workerFromDB.IsPending() {
		Fatal(t, "expected worker to be pending")
	}

	if workerFromDB.IsApproved() {
		Fatal(t, "expected worker to be pending, not pending")
	}

	if workerFromDB.IsRejected() {
		Fatal(t, "expected worker to be pending, not rejected")
	}

	if workerFromDB.ID() != workerID {
		Fatal(t, "invalid worker id")
	}

	if workerFromDB.Name() != "test_worker" {
		Fatal(t, "invalid worker name")
	}

	if workerFromDB.ColonyID() != colony.ID() {
		Fatal(t, "invalid worker colony id")
	}

	if workerFromDB.CPU() != "AMD Ryzen 9 5950X (32) @ 3.400GHz" {
		Fatal(t, "invalid worker cpu")
	}

	if workerFromDB.Cores() != 32 {
		Fatal(t, "invalid worker cores")
	}

	if workerFromDB.Mem() != 80326 {
		Fatal(t, "invalid worker mem")
	}

	if workerFromDB.GPU() != "NVIDIA GeForce RTX 2080 Ti Rev. A" {
		Fatal(t, "invalid worker gpu")
	}

	if workerFromDB.GPUs() != 1 {
		Fatal(t, "invalid worker gpus")
	}
}

func TestAddTwoWorker(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony)
	CheckError(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	CheckError(t, err)

	workers, err := db.GetWorkers()
	CheckError(t, err)
	if len(workers) != 2 {
		Fatal(t, "invalid size of workers array, expected 2")
	}
}

func TestGetWorkerByID(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony)
	CheckError(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	CheckError(t, err)

	workerFromDB, err := db.GetWorkerByID(worker1.ID())
	CheckError(t, err)
	if workerFromDB.ID() != worker1.ID() {
		Fatal(t, "failed to get worker by id")
	}
}

func TestGetWorkerByColonyID(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony1, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	CheckError(t, err)

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

	workersInColony1, err := db.GetWorkersByColonyID(colony1.ID())
	CheckError(t, err)

	counter := 0
	for _, worker := range workersInColony1 {
		if worker.ID() == worker1ID {
			counter++
		}
		if worker.ID() == worker2ID {
			counter++
		}
	}
	if counter != 2 {
		Fatal(t, "Failed to get workers in colony 1")
	}
}

func TestApproveWorker(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name")
	CheckError(t, err)

	err = db.AddColony(colony)
	CheckError(t, err)

	workerID := core.GenerateRandomID()
	worker := core.CreateWorker(workerID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker)
	CheckError(t, err)

	if !worker.IsPending() {
		Fatal(t, "expected worker to be pending")
	}

	err = db.ApproveWorker(worker)
	CheckError(t, err)

	if worker.IsPending() {
		Fatal(t, "expected worker not to be pending")
	}

	if worker.IsRejected() {
		Fatal(t, "expected worker not to be rejected")
	}

	if !worker.IsApproved() {
		Fatal(t, "expected worker to be approved")
	}

	workerFromDB, err := db.GetWorkerByID(worker.ID())
	CheckError(t, err)
	if !workerFromDB.IsApproved() {
		Fatal(t, "expected worker to be approved")
	}

	err = db.RejectWorker(worker)
	CheckError(t, err)
	if !worker.IsRejected() {
		Fatal(t, "expected worker to be rejected")
	}

	workerFromDB, err = db.GetWorkerByID(worker.ID())
	CheckError(t, err)
	if !workerFromDB.IsRejected() {
		Fatal(t, "expected worker to be rejected")
	}
}

func TestDeleteWorkers(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony1, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	CheckError(t, err)

	err = db.AddColony(colony1)
	CheckError(t, err)

	colony2, err := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	CheckError(t, err)

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

	err = db.DeleteWorkerByID(worker2.ID())
	CheckError(t, err)

	workerFromDB, err := db.GetWorkerByID(worker2.ID())
	CheckError(t, err)
	if workerFromDB != nil {
		Fatal(t, "expected worker to be nil")
	}

	err = db.AddWorker(worker2)
	CheckError(t, err)

	workerFromDB, err = db.GetWorkerByID(worker2.ID())
	CheckError(t, err)
	if workerFromDB == nil {
		Fatal(t, "expected worker not to be nil")
	}

	err = db.DeleteWorkersByColonyID(colony1.ID())
	CheckError(t, err)

	workerFromDB, err = db.GetWorkerByID(worker1.ID())
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
