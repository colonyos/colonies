package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateRuntime(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	runtimeType := "test_runtime_type"
	name := ""
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	runtime := CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)

	assert.Equal(t, PENDING, runtime.State)
	assert.True(t, runtime.IsPending())
	assert.False(t, runtime.IsApproved())
	assert.False(t, runtime.IsRejected())
	assert.Equal(t, id, runtime.ID)
	assert.Equal(t, runtimeType, runtime.RuntimeType)
	assert.Equal(t, name, runtime.Name)
	assert.Equal(t, colonyID, runtime.ColonyID)
	assert.Equal(t, cpu, runtime.CPU)
	assert.Equal(t, cores, runtime.Cores)
	assert.Equal(t, mem, runtime.Mem)
	assert.Equal(t, gpu, runtime.GPU)
	assert.Equal(t, gpus, runtime.GPUs)
}

func TestSetRuntimeID(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	runtime := CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	runtime.SetID("test_runtimeid_set")

	assert.Equal(t, runtime.ID, "test_runtimeid_set")
}

func TestSetColonyIDonRimtime(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	runtime := CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	runtime.SetColonyID("test_colonyid_set")

	assert.Equal(t, runtime.ColonyID, "test_colonyid_set")
}

func TestRuntimeEquals(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	runtime1 := CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.True(t, runtime1.Equals(runtime1))

	runtime2 := CreateRuntime(id+"X", runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType+"X", name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name+"X", colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name, colonyID+"X", cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name, colonyID, cpu+"X", cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name, colonyID, cpu, cores+1, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem+1, gpu, gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu+"X", gpus, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	runtime2 = CreateRuntime(id, runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus+1, commissionTime, lastHeardFromTime)
	assert.False(t, runtime2.Equals(runtime1))
	assert.False(t, runtime2.Equals(nil))
}

func TestIsRuntimeArraysEqual(t *testing.T) {
	runtimeType := "test_runtime_type"
	name := "test_runtime_name"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	runtime1 := CreateRuntime(GenerateRandomID(), runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	runtime2 := CreateRuntime(GenerateRandomID(), runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	runtime3 := CreateRuntime(GenerateRandomID(), runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)
	runtime4 := CreateRuntime(GenerateRandomID(), runtimeType, name, colonyID, cpu, cores, mem, gpu, gpus, commissionTime, lastHeardFromTime)

	var runtimes1 []*Runtime
	runtimes1 = append(runtimes1, runtime1)
	runtimes1 = append(runtimes1, runtime2)
	runtimes1 = append(runtimes1, runtime3)

	var runtimes2 []*Runtime
	runtimes2 = append(runtimes2, runtime2)
	runtimes2 = append(runtimes2, runtime3)
	runtimes2 = append(runtimes2, runtime1)

	var runtimes3 []*Runtime
	runtimes3 = append(runtimes3, runtime2)
	runtimes3 = append(runtimes3, runtime3)
	runtimes3 = append(runtimes3, runtime4)

	var runtimes4 []*Runtime

	assert.True(t, IsRuntimeArraysEqual(runtimes1, runtimes1))
	assert.True(t, IsRuntimeArraysEqual(runtimes1, runtimes2))
	assert.False(t, IsRuntimeArraysEqual(runtimes1, runtimes3))
	assert.False(t, IsRuntimeArraysEqual(runtimes1, runtimes4))
	assert.True(t, IsRuntimeArraysEqual(runtimes4, runtimes4))
	assert.True(t, IsRuntimeArraysEqual(nil, nil))
	assert.False(t, IsRuntimeArraysEqual(nil, runtimes2))
}

func TestRuntimeToJSON(t *testing.T) {
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	runtime1 := CreateRuntime("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_runtime_type", "test_runtime_name", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, commissionTime, lastHeardFromTime)

	jsonString, err := runtime1.ToJSON()
	assert.Nil(t, err)

	runtime2, err := ConvertJSONToRuntime(jsonString + "error")
	assert.NotNil(t, err)

	runtime2, err = ConvertJSONToRuntime(jsonString)
	assert.Nil(t, err)
	assert.True(t, runtime2.Equals(runtime1))
}

func TestRuntimeToJSONArray(t *testing.T) {
	commissionTime := time.Now()
	lastHeardFromTime := time.Now()

	var runtimes1 []*Runtime
	runtimes1 = append(runtimes1, CreateRuntime(GenerateRandomID(), "test_runtime_type", "test_runtime_name", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, commissionTime, lastHeardFromTime))
	runtimes1 = append(runtimes1, CreateRuntime(GenerateRandomID(), "test_runtime_type", "test_runtime_name", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, commissionTime, lastHeardFromTime))

	jsonString, err := ConvertRuntimeArrayToJSON(runtimes1)
	assert.Nil(t, err)

	runtimes2, err := ConvertJSONToRuntimeArray(jsonString + "error")
	assert.NotNil(t, err)

	runtimes2, err = ConvertJSONToRuntimeArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsRuntimeArraysEqual(runtimes1, runtimes2))
}
