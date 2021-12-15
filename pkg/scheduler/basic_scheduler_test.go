package scheduler

import (
	"colonies/pkg/core"
	. "colonies/pkg/utils"
	"testing"
	"time"
)

func TestCreateTask(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	task2 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	task3 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Task{task1, task2, task3}

	scheduler := CreateBasicScheduler()
	selectedTask := scheduler.Select("workerid_1", candidates)
	if selectedTask.ID() != task1.ID() {
		Fatal(t, "not the latest task was selected")
	}
}

func TestCreateTask2(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task1.SetSubmissionTime(startTime.Add(60 * time.Millisecond))

	task2 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	task3 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Task{task1, task2, task3}

	scheduler := CreateBasicScheduler()
	selectedTask := scheduler.Select("workerid_1", candidates)
	if selectedTask.ID() != task3.ID() {
		Fatal(t, "not the latest task was selected")
	}
}

func TestCreateTask3(t *testing.T) { // Note same submission times
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task1.SetSubmissionTime(startTime)

	task2 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task2.SetSubmissionTime(startTime)

	task3 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task3.SetSubmissionTime(startTime)

	candidates := []*core.Task{task1, task2, task3}

	scheduler := CreateBasicScheduler()
	selectedTask := scheduler.Select("workerid_1", candidates)
	if selectedTask.ID() != task1.ID() { // The first task will be selected
		Fatal(t, "not the latest task was selected")
	}
}

func TestCreateTask4(t *testing.T) { // Note same submission times
	candidates := []*core.Task{}

	scheduler := CreateBasicScheduler()
	selectedTask := scheduler.Select("workerid_1", candidates)
	if selectedTask != nil {
		Fatal(t, "expected nil")
	}
}

func TestCreateTask5(t *testing.T) {
	startTime := time.Now()

	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name")

	task1 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task1.SetSubmissionTime(startTime.Add(600 * time.Millisecond))

	task2 := core.CreateTask(colony.ID(), []string{"workerid_1"}, "dummy", -1, 3, 1000, 10, 1)
	task2.SetSubmissionTime(startTime.Add(100 * time.Millisecond))

	task3 := core.CreateTask(colony.ID(), []string{}, "dummy", -1, 3, 1000, 10, 1)
	task3.SetSubmissionTime(startTime.Add(300 * time.Millisecond))

	candidates := []*core.Task{task1, task2, task3}

	scheduler := CreateBasicScheduler()
	selectedTask := scheduler.Select("workerid_1", candidates)
	if selectedTask.ID() != task2.ID() {
		Fatal(t, "not the latest task was selected")
	}
}
