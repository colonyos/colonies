package core

import (
	. "colonies/pkg/utils"
	"testing"
)

func TestCreateWorker(t *testing.T) {
	worker := CreateWorker("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_worker", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	if !worker.IsPending() {
		Fatal(t, "expected worker to be pending")
	}

	if worker.IsApproved() {
		Fatal(t, "expected worker to be pending, not pending")
	}

	if worker.IsRejected() {
		Fatal(t, "expected worker to be pending, not rejected")
	}

	if worker.ID() != "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb" {
		Fatal(t, "invalid worker id")
	}

	if worker.Name() != "test_worker" {
		Fatal(t, "invalid worker name")
	}

	if worker.ColonyID() != "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834" {
		Fatal(t, "invalid worker colony id")
	}

	if worker.CPU() != "AMD Ryzen 9 5950X (32) @ 3.400GHz" {
		Fatal(t, "invalid worker cpu")
	}

	if worker.Cores() != 32 {
		Fatal(t, "invalid worker cores")
	}

	if worker.Mem() != 80326 {
		Fatal(t, "invalid worker mem")
	}

	if worker.GPU() != "NVIDIA GeForce RTX 2080 Ti Rev. A" {
		Fatal(t, "invalid worker gpu")
	}

	if worker.GPUs() != 1 {
		Fatal(t, "invalid worker gpus")
	}
}
