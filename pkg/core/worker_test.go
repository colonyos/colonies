package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWorker(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	name := "test_worker"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	worker := CreateWorker(id, name, colonyID, cpu, cores, mem, gpu, gpus)

	assert.Equal(t, PENDING, worker.Status())
	assert.True(t, worker.IsPending())
	assert.False(t, worker.IsApproved())
	assert.False(t, worker.IsRejected())
	assert.Equal(t, id, worker.ID())
	assert.Equal(t, name, worker.Name())
	assert.Equal(t, colonyID, worker.ColonyID())
	assert.Equal(t, cpu, worker.CPU())
	assert.Equal(t, cores, worker.Cores())
	assert.Equal(t, mem, worker.Mem())
	assert.Equal(t, gpu, worker.GPU())
	assert.Equal(t, gpus, worker.GPUs())
}

func TestWorkerToJSON(t *testing.T) {
	worker := CreateWorker("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_worker", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	jsonString, err := worker.ToJSON()
	assert.Nil(t, err)

	worker2, err := CreateWorkerFromJSON(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, worker.ID(), worker2.ID())
	assert.Equal(t, worker.Name(), worker2.Name())
	assert.Equal(t, worker.ColonyID(), worker2.ColonyID())
	assert.Equal(t, worker.CPU(), worker2.CPU())
	assert.Equal(t, worker.Cores(), worker2.Cores())
	assert.Equal(t, worker.Mem(), worker2.Mem())
	assert.Equal(t, worker.GPU(), worker2.GPU())
	assert.Equal(t, worker.GPUs(), worker2.GPUs())
	assert.Equal(t, worker.Status(), worker2.Status())
}
