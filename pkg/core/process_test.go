package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProcess(t *testing.T) {
	colonyID := GenerateRandomID()
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec := CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
	process := CreateProcess(processSpec)
	assert.True(t, process.ProcessSpec.Equals(processSpec))
}

func TestAssignProcess(t *testing.T) {
	colonyID := GenerateRandomID()
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec := CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
	process := CreateProcess(processSpec)

	assert.False(t, process.IsAssigned)
	process.Assign()
	assert.True(t, process.IsAssigned)
	process.Unassign()
	assert.False(t, process.IsAssigned)
}

func TestTimeCalc(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec := CreateProcessSpec(colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
	process := CreateProcess(processSpec)
	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))
	assert.False(t, process.WaitingTime() < 900000000 && process.WaitingTime() > 1200000000)
	assert.False(t, process.WaitingTime() < 3000000000 && process.WaitingTime() > 4000000000)
}

func TestProcessToJSON(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec := CreateProcessSpec(colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
	process := CreateProcess(processSpec)
	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))
	attribute1ID := GenerateRandomID()
	attribute2ID := GenerateRandomID()
	attribute3ID := GenerateRandomID()
	attribute4ID := GenerateRandomID()
	attribute5ID := GenerateRandomID()
	attribute6ID := GenerateRandomID()
	var attributes []*Attribute
	attributes = append(attributes, CreateAttribute(attribute1ID, IN, "in_key_1", "in_value_1"))
	attributes = append(attributes, CreateAttribute(attribute2ID, IN, "in_key_2", "in_value_2"))
	attributes = append(attributes, CreateAttribute(attribute3ID, ERR, "err_key_1", "err_value_1"))
	attributes = append(attributes, CreateAttribute(attribute4ID, ERR, "err_key_2", "err_value_2"))
	attributes = append(attributes, CreateAttribute(attribute5ID, OUT, "out_key_1", "out_value_1"))
	attributes = append(attributes, CreateAttribute(attribute6ID, OUT, "out_key_2", "out_value_2"))
	process.SetAttributes(attributes)

	jsonString, err := process.ToJSON()
	assert.Nil(t, err)

	process2, err := ConvertJSONToProcess(jsonString)
	assert.Nil(t, err)
	assert.True(t, process.Equals(process2))
}

func TestProcessArrayToJSON(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec1 := CreateProcessSpec(colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
	process1 := CreateProcess(processSpec1)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(1 * time.Second))
	process1.SetEndTime(startTime.Add(4 * time.Second))
	attribute1ID := GenerateRandomID()
	attribute2ID := GenerateRandomID()
	attribute3ID := GenerateRandomID()
	var attributes1 []*Attribute
	attributes1 = append(attributes1, CreateAttribute(attribute1ID, IN, "in_key_1", "in_value_1"))
	attributes1 = append(attributes1, CreateAttribute(attribute2ID, ERR, "err_key_1", "err_value_1"))
	attributes1 = append(attributes1, CreateAttribute(attribute3ID, OUT, "out_key_1", "out_value_1"))
	process1.SetAttributes(attributes1)

	processSpec2 := CreateProcessSpec(colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string))
	process2 := CreateProcess(processSpec2)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(1 * time.Second))
	process2.SetEndTime(startTime.Add(4 * time.Second))
	attribute4ID := GenerateRandomID()
	attribute5ID := GenerateRandomID()
	attribute6ID := GenerateRandomID()
	var attributes2 []*Attribute
	attributes2 = append(attributes2, CreateAttribute(attribute4ID, IN, "in_key_1", "in_value_1"))
	attributes2 = append(attributes2, CreateAttribute(attribute5ID, ERR, "err_key_1", "err_value_1"))
	attributes2 = append(attributes2, CreateAttribute(attribute6ID, OUT, "out_key_1", "out_value_1"))
	process2.SetAttributes(attributes2)

	var processes1 []*Process
	processes1 = append(processes1, process1)
	processes1 = append(processes1, process2)

	jsonString, err := ConvertProcessArrayToJSON(processes1)
	assert.Nil(t, err)
	processes2, err := ConvertJSONToProcessArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsProcessArrayEqual(processes1, processes2))
}
