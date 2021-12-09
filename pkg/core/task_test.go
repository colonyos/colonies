package core

import (
	. "colonies/pkg/utils"
	"testing"
	"time"
)

func TestCreateTask(t *testing.T) {
	colonyID := GenerateRandomID()
	task := CreateTask(colonyID, []string{"4cbb01dd59506d39f08abde667787d9d1788fb68d3156266f68773d056e820d", "37751eac5c5daa9d1842b76b3a0794b2603c4dc400547e86478bcdad912faba"}, "dummy", -1, 3, 1000, 10, 1)

	if task.TargetColonyID() != colonyID {
		Fatal(t, "invalid task id")
	}

	counter := 0
	for _, targetWorkerID := range task.TargetWorkerIDs() {
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

	if task.WorkerType() != "dummy" {
		Fatal(t, "invalid worker type in task")
	}

	if task.Timeout() != -1 {
		Fatal(t, "invalid timeout in task")
	}

	if task.MaxRetries() != 3 {
		Fatal(t, "invalid max retries in task")
	}

	if task.Mem() != 1000 {
		Fatal(t, "invalid mem in task")
	}

	if task.Cores() != 10 {
		Fatal(t, "invalid cores in task")
	}

	if task.GPUs() != 1 {
		Fatal(t, "invalid gpus in task")
	}

	if task.Assigned() == true {
		Fatal(t, "expected a new task to be unassigned")
	}

	task.Assign()

	if task.Assigned() == false {
		Fatal(t, "expected a new task to be assigned")
	}

	task.Unassign()

	if task.Assigned() == true {
		Fatal(t, "expected a new task to be unassigned after calling the Unassign function")
	}
}

func TestTimeCalc(t *testing.T) {
	colonyID := GenerateRandomID()
	task := CreateTask(colonyID, []string{}, "dummy", -1, 3, 1000, 10, 1)

	startTime := time.Now()

	task.SetSubmissionTime(startTime)
	task.SetStartTime(startTime.Add(1 * time.Second))
	task.SetEndTime(startTime.Add(4 * time.Second))

	if task.WaitingTime() < 900000000 && task.WaitingTime() > 1200000000 {
		Fatal(t, "invalid waiting time")
	}

	if task.WaitingTime() < 3000000000 && task.WaitingTime() > 4000000000 {
		Fatal(t, "invalid processing time")
	}
}
