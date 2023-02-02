package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateExecutor(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := ""
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)

	assert.Equal(t, PENDING, executor.State)
	assert.True(t, executor.IsPending())
	assert.False(t, executor.IsApproved())
	assert.False(t, executor.IsRejected())
	assert.Equal(t, id, executor.ID)
	assert.Equal(t, executorType, executor.Type)
	assert.Equal(t, name, executor.Name)
	assert.Equal(t, colonyID, executor.ColonyID)
	assert.Equal(t, cpu, executor.CPU)
	assert.Equal(t, cores, executor.Cores)
	assert.Equal(t, mem, executor.Mem)
	assert.Equal(t, gpu, executor.GPU)
	assert.Equal(t, gpus, executor.GPUs)
}

func TestSetExecutorID(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	executor.SetID("test_executor_id_set")

	assert.Equal(t, executor.ID, "test_executor_id_set")
}

func TestSetColonyIDonRimtime(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor := CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	executor.SetColonyID("test_colonyid_set")

	assert.Equal(t, executor.ColonyID, "test_colonyid_set")
}

func TestExecutorEquals(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.True(t, executor1.Equals(executor1))

	executor2 := CreateExecutor(id+"X", executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType+"X", name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name+"X", colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyID+"X", cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyID, cpu+"X", cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyID, cpu, cores+1, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem+1, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu+"X", gpus, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	executor2 = CreateExecutor(id, executorType, name, colonyID, cpu, cores, mem, gpu, gpus+1, commissionTime, lastHeardFromTime)
	assert.False(t, executor2.Equals(executor1))
	assert.False(t, executor2.Equals(nil))
}

func TestIsExecutorArraysEqual(t *testing.T) {
	executorType := "test_executor_type"
	name := "test_executor_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor(GenerateRandomID(), executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	executor2 := CreateExecutor(GenerateRandomID(), executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	executor3 := CreateExecutor(GenerateRandomID(), executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	executor4 := CreateExecutor(GenerateRandomID(), executorType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)

	var executors1 []*Executor
	executors1 = append(executors1, executor1)
	executors1 = append(executors1, executor2)
	executors1 = append(executors1, executor3)

	var executors2 []*Executor
	executors2 = append(executors2, executor2)
	executors2 = append(executors2, executor3)
	executors2 = append(executors2, executor1)

	var executors3 []*Executor
	executors3 = append(executors3, executor2)
	executors3 = append(executors3, executor3)
	executors3 = append(executors3, executor4)

	var executors4 []*Executor

	assert.True(t, IsExecutorArraysEqual(executors1, executors1))
	assert.True(t, IsExecutorArraysEqual(executors1, executors2))
	assert.False(t, IsExecutorArraysEqual(executors1, executors3))
	assert.False(t, IsExecutorArraysEqual(executors1, executors4))
	assert.True(t, IsExecutorArraysEqual(executors4, executors4))
	assert.True(t, IsExecutorArraysEqual(nil, nil))
	assert.False(t, IsExecutorArraysEqual(nil, executors2))
}

func TestExecutorToJSON(t *testing.T) {
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	executor1 := CreateExecutor("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_executor_type", "test_executor_name", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, commissionTime, lastHeardFromTime)

	jsonString, err := executor1.ToJSON()
	assert.Nil(t, err)

	executor2, err := ConvertJSONToExecutor(jsonString + "error")
	assert.NotNil(t, err)

	executor2, err = ConvertJSONToExecutor(jsonString)
	assert.Nil(t, err)
	assert.True(t, executor2.Equals(executor1))
}

func TestExecutorToJSONArray(t *testing.T) {
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	var executors1 []*Executor
	executors1 = append(executors1, CreateExecutor(GenerateRandomID(), "test_executor_type", "test_executor_name", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, commissionTime, lastHeardFromTime))
	executors1 = append(executors1, CreateExecutor(GenerateRandomID(), "test_executor_type", "test_executor_name", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, commissionTime, lastHeardFromTime))

	jsonString, err := ConvertExecutorArrayToJSON(executors1)
	assert.Nil(t, err)

	executors2, err := ConvertJSONToExecutorArray(jsonString + "error")
	assert.NotNil(t, err)

	executors2, err = ConvertJSONToExecutorArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsExecutorArraysEqual(executors1, executors2))
}
