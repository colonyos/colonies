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

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process := CreateProcess(processSpec)
	assert.True(t, process.ProcessSpec.Equals(processSpec))
}

func TestCreateProcessFromDB(t *testing.T) {
	colonyID := GenerateRandomID()
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	var attributes []*Attribute

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process := CreateProcessFromDB(processSpec, GenerateRandomID(), GenerateRandomID(), true, FAILED, time.Now(), time.Now(), time.Now(), time.Now(), 2, attributes)
	assert.True(t, process.Equals(process))
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

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process := CreateProcess(processSpec)

	assert.False(t, process.IsAssigned)
	process.Assign()
	assert.True(t, process.IsAssigned)
	process.Unassign()
	assert.False(t, process.IsAssigned)
}

func TestProcessTimeCalc(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process := CreateProcess(processSpec)
	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))
	assert.False(t, process.WaitingTime() < 900000000 && process.WaitingTime() > 1200000000)
	assert.False(t, process.WaitingTime() < 3000000000 && process.WaitingTime() > 4000000000)
}

func TestProcessEquals(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	processSpec1 := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process1 := CreateProcess(processSpec1)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(1 * time.Second))
	process1.SetEndTime(startTime.Add(4 * time.Second))
	assert.True(t, process1.Equals(process1))
	assert.False(t, process1.Equals(nil))

	colonyID2 := GenerateRandomID()
	processSpec2 := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID2, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)

	process2 := CreateProcess(processSpec2)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(1 * time.Second))
	process2.SetEndTime(startTime.Add(4 * time.Second))

	assert.False(t, process1.Equals(process2))
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

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{"test_name_2"}, 1)
	process := CreateProcess(processSpec)
	process.AddParent(GenerateRandomID())
	process.AddParent(GenerateRandomID())
	process.SetProcessGraphID(GenerateRandomID())
	process.AddChild(GenerateRandomID())
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
	attributes = append(attributes, CreateAttribute(attribute1ID, GenerateRandomID(), IN, "in_key_1", "in_value_1"))
	attributes = append(attributes, CreateAttribute(attribute2ID, GenerateRandomID(), IN, "in_key_2", "in_value_2"))
	attributes = append(attributes, CreateAttribute(attribute3ID, GenerateRandomID(), ERR, "err_key_1", "err_value_1"))
	attributes = append(attributes, CreateAttribute(attribute4ID, GenerateRandomID(), ERR, "err_key_2", "err_value_2"))
	attributes = append(attributes, CreateAttribute(attribute5ID, GenerateRandomID(), OUT, "out_key_1", "out_value_1"))
	attributes = append(attributes, CreateAttribute(attribute6ID, GenerateRandomID(), OUT, "out_key_2", "out_value_2"))
	process.SetAttributes(attributes)

	jsonString, err := process.ToJSON()
	assert.Nil(t, err)

	process2, err := ConvertJSONToProcess(jsonString + "error")
	assert.NotNil(t, err)

	process2, err = ConvertJSONToProcess(jsonString)
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

	processSpec1 := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process1 := CreateProcess(processSpec1)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(1 * time.Second))
	process1.SetEndTime(startTime.Add(4 * time.Second))
	attribute1ID := GenerateRandomID()
	attribute2ID := GenerateRandomID()
	attribute3ID := GenerateRandomID()
	var attributes1 []*Attribute
	attributes1 = append(attributes1, CreateAttribute(attribute1ID, GenerateRandomID(), IN, "in_key_1", "in_value_1"))
	attributes1 = append(attributes1, CreateAttribute(attribute2ID, GenerateRandomID(), ERR, "err_key_1", "err_value_1"))
	attributes1 = append(attributes1, CreateAttribute(attribute3ID, GenerateRandomID(), OUT, "out_key_1", "out_value_1"))
	process1.SetAttributes(attributes1)

	processSpec2 := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process2 := CreateProcess(processSpec2)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(1 * time.Second))
	process2.SetEndTime(startTime.Add(4 * time.Second))
	attribute4ID := GenerateRandomID()
	attribute5ID := GenerateRandomID()
	attribute6ID := GenerateRandomID()
	var attributes2 []*Attribute
	attributes2 = append(attributes2, CreateAttribute(attribute4ID, GenerateRandomID(), IN, "in_key_1", "in_value_1"))
	attributes2 = append(attributes2, CreateAttribute(attribute5ID, GenerateRandomID(), ERR, "err_key_1", "err_value_1"))
	attributes2 = append(attributes2, CreateAttribute(attribute6ID, GenerateRandomID(), OUT, "out_key_1", "out_value_1"))
	process2.SetAttributes(attributes2)

	var processes1 []*Process
	processes1 = append(processes1, process1)
	processes1 = append(processes1, process2)

	jsonString, err := ConvertProcessArrayToJSON(processes1)
	assert.Nil(t, err)

	processes2, err := ConvertJSONToProcessArray(jsonString + "error")
	assert.NotNil(t, err)

	processes2, err = ConvertJSONToProcessArray(jsonString)
	assert.Nil(t, err)
	assert.True(t, IsProcessArraysEqual(processes1, processes2))
}

func TestProcessingTime(t *testing.T) {
	colonyID := GenerateRandomID()
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	var attributes []*Attribute

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, make(map[string]string), []string{}, 1)
	process := CreateProcessFromDB(processSpec, GenerateRandomID(), GenerateRandomID(), true, RUNNING, time.Time{}, time.Time{}, time.Time{}, time.Time{}, 2, attributes)

	processingTime := int64(process.ProcessingTime())
	assert.True(t, processingTime > 0)

	process.SetState(WAITING)
	processingTime = int64(process.ProcessingTime())
	assert.True(t, processingTime == 0)
}
