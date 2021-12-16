package database

import (
	"colonies/pkg/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddTask(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	worker1ID := core.GenerateRandomID()
	worker2ID := core.GenerateRandomID()

	task := core.CreateTask(colonyID, []string{worker1ID, worker2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.Equal(t, colonyID, taskFromDB.TargetColonyID())
	assert.Contains(t, taskFromDB.TargetWorkerIDs(), worker1ID)
	assert.Contains(t, taskFromDB.TargetWorkerIDs(), worker2ID)
}

func TestDeleteTasks(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()
	worker1ID := core.GenerateRandomID()
	worker2ID := core.GenerateRandomID()

	task1 := core.CreateTask(colonyID, []string{worker1ID, worker2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	task2 := core.CreateTask(colonyID, []string{worker1ID, worker2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	assert.Nil(t, err)

	task3 := core.CreateTask(colonyID, []string{worker1ID, worker2ID}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task3)
	assert.Nil(t, err)

	numberOfTasks, err := db.NumberOfTasks()
	assert.Nil(t, err)
	assert.Equal(t, 3, numberOfTasks)

	err = db.DeleteTaskByID(task1.ID())
	assert.Nil(t, err)

	numberOfTasks, err = db.NumberOfTasks()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfTasks)

	err = db.DeleteAllTasks()
	assert.Nil(t, err)

	numberOfTasks, err = db.NumberOfTasks()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfTasks)
}

func TestDeleteAllTasksAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()

	task1 := core.CreateTask(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(task1.ID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteAllTasks()
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(task1.ID(), "test_key1", core.IN)
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)
}

func TestDeleteTasksAndAttributes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colonyID := core.GenerateRandomID()

	task1 := core.CreateTask(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	task2 := core.CreateTask(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(task1.ID(), core.IN, "test_key1", "test_value1")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	attribute = core.CreateAttribute(task2.ID(), core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	err = db.DeleteTaskByID(task1.ID())
	assert.Nil(t, err)

	attributeFromDB, err := db.GetAttribute(task1.ID(), "test_key1", core.IN)
	assert.Nil(t, err)
	assert.Nil(t, attributeFromDB)

	attributeFromDB, err = db.GetAttribute(task2.ID(), "test_key2", core.IN)
	assert.Nil(t, err)
	assert.NotNil(t, attributeFromDB) // Not deleted as it belongs to task 2
}

func TestAssign(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	assert.Nil(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, taskFromDB.Status())
	assert.False(t, taskFromDB.Assigned())

	err = db.AssignWorker(worker.ID(), task)
	assert.Nil(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.True(t, taskFromDB.Assigned())
	assert.False(t, int64(taskFromDB.StartTime().Sub(taskFromDB.SubmissionTime())) < 0)
	assert.Equal(t, core.RUNNING, taskFromDB.Status())

	err = db.UnassignWorker(task)
	assert.Nil(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	assert.Nil(t, err)
	assert.False(t, taskFromDB.Assigned())
	assert.False(t, int64(taskFromDB.EndTime().Sub(taskFromDB.StartTime())) < 0)
}

func TestMarkSuccessful(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	assert.Nil(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, task.Status())

	err = db.MarkSuccessful(task)
	assert.NotNil(t, err) // Not possible to set waiting task to successfull

	err = db.AssignWorker(worker.ID(), task)
	assert.Nil(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, task.Status())

	err = db.MarkSuccessful(task)
	assert.Nil(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.SUCCESS, taskFromDB.Status())

	err = db.MarkFailed(task)
	assert.NotNil(t, err) // Not possible to set successful task as failed
}

func TestMarkFailed(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	assert.Nil(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)

	assert.Equal(t, core.WAITING, task.Status())

	err = db.MarkFailed(task)
	assert.NotNil(t, err) // Not possible to set waiting task to failed

	err = db.AssignWorker(worker.ID(), task)
	assert.Nil(t, err)

	taskFromDB, err := db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.RUNNING, taskFromDB.Status())

	err = db.MarkFailed(task)
	assert.Nil(t, err)

	taskFromDB, err = db.GetTaskByID(task.ID())
	assert.Nil(t, err)

	assert.Equal(t, core.FAILED, taskFromDB.Status())

	err = db.MarkFailed(task)
	assert.NotNil(t, err) // Not possible to set successful task as failed
}

func TestReset(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	worker := core.CreateWorker(colony.ID(), "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker)
	assert.Nil(t, err)

	task := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)
	err = db.AssignWorker(worker.ID(), task)
	assert.Nil(t, err)
	err = db.MarkFailed(task)
	assert.Nil(t, err)

	task = core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)
	err = db.AssignWorker(worker.ID(), task)
	assert.Nil(t, err)
	err = db.MarkFailed(task)
	assert.Nil(t, err)

	task = core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task)
	assert.Nil(t, err)
	err = db.AssignWorker(worker.ID(), task)
	assert.Nil(t, err)
	err = db.MarkFailed(task)
	assert.Nil(t, err)

	numberOfFailedTasks, err := db.NumberOfFailedTasks()
	assert.Equal(t, 3, numberOfFailedTasks)

	err = db.ResetTask(task)
	assert.Nil(t, err)

	numberOfFailedTasks, err = db.NumberOfFailedTasks()
	assert.Equal(t, 2, numberOfFailedTasks)

	err = db.ResetAllTasks(task)
	assert.Nil(t, err)

	numberOfFailedTasks, err = db.NumberOfFailedTasks()
	assert.Equal(t, 0, numberOfFailedTasks)
}

func TestSearchTask1(t *testing.T) {
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

	task1 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	assert.Nil(t, err)

	tasksFromDB, err := db.SearchTasks(colony.ID(), worker2.ID())
	assert.Nil(t, err)

	assert.Len(t, tasksFromDB, 1)
	assert.Equal(t, task1.ID(), tasksFromDB[0].ID())
}

func TestSearchTask2(t *testing.T) {
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

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	assert.Nil(t, err)

	task3 := core.CreateTask(colony.ID(), []string{worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task3)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	tasksFromDB, err := db.SearchTasks(colony.ID(), worker2.ID())
	assert.Nil(t, err)
	assert.Len(t, tasksFromDB, 2)

	counter := 0
	for _, taskFromDB := range tasksFromDB {
		if taskFromDB.ID() == task1.ID() {
			counter++
		}

		if taskFromDB.ID() == task2.ID() {
			counter++
		}
	}

	assert.Equal(t, 2, counter)
}

func TestSearchTask3(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	assert.Nil(t, err)
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

	// Here, we are testing that the order of targetWorkerIDs strings does not matter

	task1 := core.CreateTask(colony.ID(), []string{worker1.ID(), worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{worker1.ID(), worker2.ID()}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	assert.Nil(t, err)

	tasksFromDB, err := db.SearchTasks(colony.ID(), worker1.ID())
	assert.Nil(t, err)

	assert.Len(t, tasksFromDB, 1)
	assert.Equal(t, task1.ID(), tasksFromDB[0].ID())

	tasksFromDB, err = db.SearchTasks(colony.ID(), worker2.ID())
	assert.Nil(t, err)

	assert.Len(t, tasksFromDB, 1)
	assert.Equal(t, task1.ID(), tasksFromDB[0].ID())
}

func TestSearchTaskAssigned(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	err = db.AddColony(colony)
	assert.Nil(t, err)

	worker1ID := core.GenerateRandomID()
	worker1 := core.CreateWorker(worker1ID, "test_worker", colony.ID(), "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)
	err = db.AddWorker(worker1)
	assert.Nil(t, err)

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task1)
	assert.Nil(t, err)

	time.Sleep(50 * time.Millisecond)

	task2 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	err = db.AddTask(task2)
	assert.Nil(t, err)

	numberOfTasks, err := db.NumberOfTasks()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfTasks)

	numberOfRunningTasks, err := db.NumberOfRunningTasks()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfRunningTasks)

	numberOfSuccesfulTasks, err := db.NumberOfSuccessfulTasks()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfSuccesfulTasks)

	numberOfFailedTasks, err := db.NumberOfFailedTasks()
	assert.Nil(t, err)
	assert.Equal(t, 0, numberOfFailedTasks)

	tasksFromDB1, err := db.SearchTasks(colony.ID(), worker1.ID())
	assert.Nil(t, err)
	assert.Equal(t, task1.ID(), tasksFromDB1[0].ID())
	assert.Len(t, tasksFromDB1, 1)

	err = db.AssignWorker(worker1.ID(), tasksFromDB1[0])
	assert.Nil(t, err)

	numberOfRunningTasks, err = db.NumberOfRunningTasks()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfRunningTasks)

	tasksFromDB2, err := db.SearchTasks(colony.ID(), worker1.ID())
	assert.Nil(t, err)
	assert.Equal(t, task2.ID(), tasksFromDB2[0].ID())

	err = db.AssignWorker(worker1.ID(), tasksFromDB2[0])
	assert.Nil(t, err)

	numberOfRunningTasks, err = db.NumberOfRunningTasks()
	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfRunningTasks)

	err = db.MarkSuccessful(tasksFromDB1[0])
	assert.Nil(t, err)

	err = db.MarkFailed(tasksFromDB2[0])
	assert.Nil(t, err)

	numberOfSuccesfulTasks, err = db.NumberOfSuccessfulTasks()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfSuccesfulTasks)

	numberOfFailedTasks, err = db.NumberOfFailedTasks()
	assert.Nil(t, err)
	assert.Equal(t, 1, numberOfFailedTasks)
}
