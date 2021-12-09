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

	colonyID := core.GenerateRandomID()
	worker1ID := core.GenerateRandomID()
	worker2ID := core.GenerateRandomID()

	task := core.CreateTask(colonyID, []string{worker1ID, worker2ID}, "dummy", -1, 3, 1000, 10, 1)

	err = db.AddTask(task)
	CheckError(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.TargetColonyID() != colonyID {
		Fatal(t, "invalid task id")
	}

	counter := 0
	for _, targetWorkerID := range taskFromDB.TargetWorkerIDs() {
		if targetWorkerID == worker1ID {
			counter++
		}

		if targetWorkerID == worker2ID {
			counter++
		}
	}

	if counter != 2 {
		Fatal(t, "invalid target worker ids in task")
	}
}

func TestAssign(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name")
	CheckError(t, err)

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	CheckError(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	CheckError(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Status() != core.WAITING {
		Fatal(t, "expected task status to be running")
	}

	if taskFromDB.Assigned() == true {
		Fatal(t, "expected new task to be unassigned")
	}

	err = db.AssignWorker(worker.ID(), task)
	CheckError(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Assigned() == false {
		Fatal(t, "expected task to be assigned")
	}

	if int64(taskFromDB.StartTime().Sub(taskFromDB.SubmissionTime())) < 0 {
		Fatal(t, "incorrect start or end time")
	}

	if taskFromDB.Status() != core.RUNNING {
		Fatal(t, "expected task status to be running")
	}

	err = db.UnassignWorker(task)
	CheckError(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Assigned() == true {
		Fatal(t, "expected task to be unassigned")
	}

	if int64(taskFromDB.EndTime().Sub(taskFromDB.StartTime())) < 0 {
		Fatal(t, "incorrect start or end time")
	}
}

func TestMarkSuccessful(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name")
	CheckError(t, err)

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	CheckError(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	CheckError(t, err)

	if task.Status() != core.WAITING {
		Fatal(t, "expected task status to be running")
	}

	err = db.MarkSuccessful(task)
	if err == nil { // Not possible to set waiting task to successfull
		t.Fatal(err)
	}

	err = db.AssignWorker(worker.ID(), task)
	CheckError(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Status() != core.RUNNING {
		Fatal(t, "expected task status to be running")
	}

	err = db.MarkSuccessful(task)
	CheckError(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Status() != core.SUCCESS {
		Fatal(t, "expected task status to be successful")
	}

	err = db.MarkFailed(task)
	if err == nil { // Not possible to set successful task as failed
		t.Fatal(err)
	}
}

func TestMarkFailed(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name")
	CheckError(t, err)

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	CheckError(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	CheckError(t, err)

	if task.Status() != core.WAITING {
		Fatal(t, "expected task status to be running")
	}

	err = db.MarkFailed(task)
	if err == nil { // Not possible to set waiting task to failed
		t.Fatal(err)
	}

	err = db.AssignWorker(worker.ID(), task)
	CheckError(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Status() != core.RUNNING {
		Fatal(t, "expected task status to be running")
	}

	err = db.MarkFailed(task)
	CheckError(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	CheckError(t, err)

	if taskFromDB.Status() != core.FAILED {
		Fatal(t, "expected task status to be failed")
	}

	err = db.MarkFailed(task)
	if err == nil { // Not possible to set successful task as failed
		t.Fatal(err)
	}
}

func TestSearchTask1(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name_1")
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

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
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

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	CheckError(t, err)

	worker2ID := core.GenerateRandomID()
	worker2 := core.CreateWorker(worker2ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
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

func TestSearchTaskAssigned(t *testing.T) {
	db, err := PrepareTests()
	CheckError(t, err)

	colony, err := core.CreateColony("test_colony_name_1")
	CheckError(t, err)
	err = db.AddColony(colony)
	CheckError(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	CheckError(t, err)

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
}
