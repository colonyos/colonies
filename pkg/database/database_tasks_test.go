package database

import (
	"colonies/pkg/core"
	. "colonies/pkg/utils"
	"testing"
	"time"
)

func TestAddTask(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	task := core.CreateTask("6cee1e51cf19fad8ac9deb8e61cfc301009d6e4153fe383e9abcd6f9f1896df5", []string{"4cbb01dd59506d39f08abde667787d9d1788fb68d3156266f68773d056e820d", "37751eac5c5daa9d1842b76b3a0794b2603c4dc400547e86478bcdad912faba"}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	CheckError(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.TargetColonyID() != "6cee1e51cf19fad8ac9deb8e61cfc301009d6e4153fe383e9abcd6f9f1896df5" {
		Fatal(t, "invalid task id")
	}

	counter := 0
	for _, targetWorkerID := range taskFromDB.TargetWorkerIDs() {
		if targetWorkerID == "4cbb01dd59506d39f08abde667787d9d1788fb68d3156266f68773d056e820d" {
			counter++
		}

		if targetWorkerID == "37751eac5c5daa9d1842b76b3a0794b2603c4dc400547e86478bcdad912faba" {
			counter++
		}
	}

	if counter != 2 {
		Fatal(t, "invalid target worker ids in task")
	}
}

func TestSearchTask1(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)
	err = db.AddColony(colony)
	CheckError(t, err)

	worker1 := core.CreateWorker("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2 := core.CreateWorker("5dfda4f1d4be06bf9d9a143737fc87698e65f09c404c05de80ed43a49fe9aea", "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker2)
	CheckError(t, err)

	task1 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	CheckError(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	CheckError(t, err)

	tasksFromDB, err := db.SearchTask(colony.ID(), worker2.ID())
	CheckError(t, err)

	if len(tasksFromDB) > 1 {
		Fatal(t, "expected one task")
	}

	if tasksFromDB[0].ID() != task1.ID() {
		Fatal(t, "expected task 1")
	}
}

func TestSearchTask2(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)
	err = db.AddColony(colony)
	CheckError(t, err)

	worker1 := core.CreateWorker("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2 := core.CreateWorker("5dfda4f1d4be06bf9d9a143737fc87698e65f09c404c05de80ed43a49fe9aea", "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker2)
	CheckError(t, err)

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	CheckError(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	CheckError(t, err)

	task3 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task3)
	CheckError(t, err)

	time.Sleep(50 * time.Millisecond)

	tasksFromDB, err := db.SearchTask(colony.ID(), worker2.ID())
	CheckError(t, err)

	if len(tasksFromDB) != 2 {
		Fatal(t, "expected two tasks")
	}

	counter := 0
	for _, taskFromDB := range tasksFromDB {
		if taskFromDB.ID() == task1.ID() {
			counter++
		}

		if taskFromDB.ID() == task2.ID() {
			counter++
		}
	}

	if counter != 2 {
		Fatal(t, "expected two tasks")
	}
}

func TestSearchTask3(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)
	err = db.AddColony(colony)
	CheckError(t, err)

	worker1 := core.CreateWorker("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2 := core.CreateWorker("5dfda4f1d4be06bf9d9a143737fc87698e65f09c404c05de80ed43a49fe9aea", "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker2)
	CheckError(t, err)

	// Here, we are testing that the order of targetWorkerIDs strings does not matter

	task1 := core.CreateTask(colony.ID(), []string{worker1.ID(), worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	CheckError(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{worker1.ID(), worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	CheckError(t, err)

	tasksFromDB, err := db.SearchTask(colony.ID(), worker1.ID())
	CheckError(t, err)

	if len(tasksFromDB) != 1 {
		Fatal(t, "expected one task")
	}

	if tasksFromDB[0].ID() != task1.ID() {
		Fatal(t, "expected task 1")
	}

	tasksFromDB, err = db.SearchTask(colony.ID(), worker2.ID())
	CheckError(t, err)

	if len(tasksFromDB) != 1 {
		Fatal(t, "expected one task")
	}

	if tasksFromDB[0].ID() != task1.ID() {
		Fatal(t, "expected task 1")
	}
}
