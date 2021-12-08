package core

import (
	. "colonies/pkg/utils"
	"testing"
)

func TestCreateTask(t *testing.T) {
	task := CreateTask("6cee1e51cf19fad8ac9deb8e61cfc301009d6e4153fe383e9abcd6f9f1896df5", []string{"4cbb01dd59506d39f08abde667787d9d1788fb68d3156266f68773d056e820d", "37751eac5c5daa9d1842b76b3a0794b2603c4dc400547e86478bcdad912faba"}, "dummy", -1, 3, 1000, 10, 1)

	if task.TargetColonyID() != "6cee1e51cf19fad8ac9deb8e61cfc301009d6e4153fe383e9abcd6f9f1896df5" {
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
}
