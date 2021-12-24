package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateComputer(t *testing.T) {
	id := "1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb"
	name := "test_computer"
	colonyID := "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834"
	cpu := "AMD Ryzen 9 5950X (32) @ 3.400GHz"
	cores := 32
	mem := 80326
	gpu := "NVIDIA GeForce RTX 2080 Ti Rev. A"
	gpus := 1

	computer := CreateComputer(id, name, colonyID, cpu, cores, mem, gpu, gpus)

	assert.Equal(t, PENDING, computer.Status)
	assert.True(t, computer.IsPending())
	assert.False(t, computer.IsApproved())
	assert.False(t, computer.IsRejected())
	assert.Equal(t, id, computer.ID)
	assert.Equal(t, name, computer.Name)
	assert.Equal(t, colonyID, computer.ColonyID)
	assert.Equal(t, cpu, computer.CPU)
	assert.Equal(t, cores, computer.Cores)
	assert.Equal(t, mem, computer.Mem)
	assert.Equal(t, gpu, computer.GPU)
	assert.Equal(t, gpus, computer.GPUs)
}

func TestComputerToJSON(t *testing.T) {
	computer1 := CreateComputer("1e1bfca6feb8a13df3cbbca1104f20b4b29c311724ee5f690356257108023fb", "test_computer", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1)

	jsonString, err := computer1.ToJSON()
	assert.Nil(t, err)

	computer2, err := ConvertJSONToComputer(jsonString)
	assert.Nil(t, err)
	assert.True(t, computer2.Equals(computer1))
}

func TestComputerToJSONArray(t *testing.T) {
	var computers1 []*Computer

	computers1 = append(computers1, CreateComputer(GenerateRandomID(), "test_computer", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1))
	computers1 = append(computers1, CreateComputer(GenerateRandomID(), "test_computer", "e0a17fead699b3e3b3eec21a3ab0efad54224f6eb22f4550abe9f2a207440834", "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1))

	jsonString, err := ConvertComputerArrayToJSON(computers1)
	assert.Nil(t, err)

	computers2, err := ConvertJSONToComputerArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsComputerArraysEqual(computers1, computers2))
}
