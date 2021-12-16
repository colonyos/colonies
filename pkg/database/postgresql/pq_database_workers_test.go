package postgresql

import (
	"colonies/pkg/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddWorker(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	workerID := core.GenerateRandomID()
	worker := core.CreateWorker(workerID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker)
	assert.Nil(t, err)

	workers, err := db.GetWorkers()
	assert.Nil(t, err)

	workerFromDB := workers[0]

	assert.True(t, workerFromDB.IsPending())
	assert.False(t, workerFromDB.IsApproved())
	assert.False(t, workerFromDB.IsRejected())
	assert.Equal(t, workerID, workerFromDB.ID())
	assert.Equal(t, "test_worker", workerFromDB.Name())
	assert.Equal(t, colony.ID(), workerFromDB.ColonyID())
	assert.Equal(t, "AMD Ryzen 9 5950X (32) @ 3.400GHz", workerFromDB.CPU())
	assert.Equal(t, 32, workerFromDB.Cores())
	assert.Equal(t, 80326, workerFromDB.Mem())
	assert.Equal(t, "NVIDIA GeForce RTX 2080 Ti Rev. A", workerFromDB.GPU())
	assert.Equal(t, 1, workerFromDB.GPUs())
}

func TestAddTwoWorker(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	assert.Nil(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	assert.Nil(t, err)

	workers, err := db.GetWorkers()
	assert.Nil(t, err)
	assert.Len(t, workers, 2)
}

func TestGetWorkerByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker1)
	assert.Nil(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker2)
	assert.Nil(t, err)

	workerFromDB, err := db.GetWorkerByID(worker1.ID())
	assert.Nil(t, err)
	assert.Equal(t, worker1.ID(), workerFromDB.ID())
}

func TestGetWorkerByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony1 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	err = db.AddColony(colony1)
	assert.Nil(t, err)

	colony2 := core.CreateColony(core.GenerateRandomID(), "test_colony_name_2")
	assert.Nil(t, err)

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

	workersInColony1, err := db.GetWorkersByColonyID(colony1.ID())
	assert.Nil(t, err)

	counter := 0
	for _, worker := range workersInColony1 {
		if worker.ID() == worker1ID {
			counter++
		}
		if worker.ID() == worker2ID {
			counter++
		}
	}

	assert.Equal(t, 2, counter)
}

func TestApproveWorker(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	err = db.AddColony(colony)
	assert.Nil(t, err)

	workerID := core.GenerateRandomID()
	worker := core.CreateWorker(workerID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	err = db.AddWorker(worker)
	assert.Nil(t, err)

	assert.True(t, worker.IsPending())

	err = db.ApproveWorker(worker)
	assert.Nil(t, err)

	assert.False(t, worker.IsPending())
	assert.False(t, worker.IsRejected())
	assert.True(t, worker.IsApproved())

	workerFromDB, err := db.GetWorkerByID(worker.ID())
	assert.Nil(t, err)
	assert.True(t, workerFromDB.IsApproved())

	err = db.RejectWorker(worker)
	assert.Nil(t, err)
	assert.True(t, worker.IsRejected())

	workerFromDB, err = db.GetWorkerByID(worker.ID())
	assert.Nil(t, err)
	assert.True(t, worker.IsRejected())
}

func TestDeleteWorkers(t *testing.T) {
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

	err = db.DeleteWorkerByID(worker2.ID())
	assert.Nil(t, err)

	workerFromDB, err := db.GetWorkerByID(worker2.ID())
	assert.Nil(t, err)
	assert.Nil(t, workerFromDB)

	err = db.AddWorker(worker2)
	assert.Nil(t, err)

	workerFromDB, err = db.GetWorkerByID(worker2.ID())
	assert.Nil(t, err)
	assert.NotNil(t, workerFromDB)

	err = db.DeleteWorkersByColonyID(colony1.ID())
	assert.Nil(t, err)

	workerFromDB, err = db.GetWorkerByID(worker1.ID())
	assert.Nil(t, err)
	assert.Nil(t, workerFromDB)

	workerFromDB, err = db.GetWorkerByID(worker2.ID())
	assert.Nil(t, err)
	assert.Nil(t, workerFromDB)

	workerFromDB, err = db.GetWorkerByID(worker3.ID())
	assert.Nil(t, err)
	assert.NotNil(t, workerFromDB)
}
