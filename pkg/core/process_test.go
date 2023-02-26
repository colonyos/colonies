package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProcess(t *testing.T) {
	colonyID := GenerateRandomID()
	executor1ID := GenerateRandomID()
	executor2ID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process := CreateProcess(funcSpec)
	assert.True(t, process.FunctionSpec.Equals(funcSpec))
}

func TestCreateProcessFromDB(t *testing.T) {
	colonyID := GenerateRandomID()
	executor1ID := GenerateRandomID()
	executor2ID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	var attributes []Attribute

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process := CreateProcessFromDB(funcSpec, GenerateRandomID(), GenerateRandomID(), true, FAILED, time.Now(), time.Now(), time.Now(), time.Now(), time.Now(), []string{"errormsg"}, 2, attributes)
	assert.True(t, process.Equals(process))
}

func TestAssignProcess(t *testing.T) {
	colonyID := GenerateRandomID()
	executor1ID := GenerateRandomID()
	executor2ID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process := CreateProcess(funcSpec)

	assert.False(t, process.IsAssigned)
	process.Assign()
	assert.True(t, process.IsAssigned)
	process.Unassign()
	assert.False(t, process.IsAssigned)
}

func TestProcessTimeCalc(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process := CreateProcess(funcSpec)
	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))
	assert.False(t, process.WaitingTime() < 900000000 && process.WaitingTime() > 1200000000)
	assert.False(t, process.WaitingTime() < 3000000000 && process.WaitingTime() > 4000000000)
}

func TestProcessEquals(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	funcSpec1 := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process1 := CreateProcess(funcSpec1)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(1 * time.Second))
	process1.SetEndTime(startTime.Add(4 * time.Second))
	process1.Input = []string{"input1"}
	process1.Output = []string{"output1"}
	assert.True(t, process1.Equals(process1))
	assert.False(t, process1.Equals(nil))

	colonyID2 := GenerateRandomID()
	funcSpec2 := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID2, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "testl_label")

	process2 := CreateProcess(funcSpec2)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(1 * time.Second))
	process2.SetEndTime(startTime.Add(4 * time.Second))

	assert.False(t, process1.Equals(process2))
}

func TestProcessToJSON(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	executorType := "test_executor_type"
	maxExecTime := -1
	maxWaitTime := -1
	maxRetries := 3

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{"test_name_2"}, 1, "test_label")
	process := CreateProcess(funcSpec)
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
	var attributes []Attribute
	attributes = append(attributes, CreateAttribute(attribute1ID, GenerateRandomID(), "", IN, "in_key_1", "in_value_1"))
	attributes = append(attributes, CreateAttribute(attribute2ID, GenerateRandomID(), GenerateRandomID(), IN, "in_key_2", "in_value_2"))
	attributes = append(attributes, CreateAttribute(attribute3ID, GenerateRandomID(), "", ERR, "err_key_1", "err_value_1"))
	attributes = append(attributes, CreateAttribute(attribute4ID, GenerateRandomID(), "", ERR, "err_key_2", "err_value_2"))
	attributes = append(attributes, CreateAttribute(attribute5ID, GenerateRandomID(), GenerateRandomID(), OUT, "out_key_1", "out_value_1"))
	attributes = append(attributes, CreateAttribute(attribute6ID, GenerateRandomID(), "", OUT, "out_key_2", "out_value_2"))
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
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	funcSpec1 := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process1 := CreateProcess(funcSpec1)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(1 * time.Second))
	process1.SetEndTime(startTime.Add(4 * time.Second))
	attribute1ID := GenerateRandomID()
	attribute2ID := GenerateRandomID()
	attribute3ID := GenerateRandomID()
	var attributes1 []Attribute
	attributes1 = append(attributes1, CreateAttribute(attribute1ID, GenerateRandomID(), "", IN, "in_key_1", "in_value_1"))
	attributes1 = append(attributes1, CreateAttribute(attribute2ID, GenerateRandomID(), GenerateRandomID(), ERR, "err_key_1", "err_value_1"))
	attributes1 = append(attributes1, CreateAttribute(attribute3ID, GenerateRandomID(), "", OUT, "out_key_1", "out_value_1"))
	process1.SetAttributes(attributes1)

	funcSpec2 := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process2 := CreateProcess(funcSpec2)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(1 * time.Second))
	process2.SetEndTime(startTime.Add(4 * time.Second))
	attribute4ID := GenerateRandomID()
	attribute5ID := GenerateRandomID()
	attribute6ID := GenerateRandomID()
	var attributes2 []Attribute
	attributes2 = append(attributes2, CreateAttribute(attribute4ID, GenerateRandomID(), "", IN, "in_key_1", "in_value_1"))
	attributes2 = append(attributes2, CreateAttribute(attribute5ID, GenerateRandomID(), "", ERR, "err_key_1", "err_value_1"))
	attributes2 = append(attributes2, CreateAttribute(attribute6ID, GenerateRandomID(), GenerateRandomID(), OUT, "out_key_1", "out_value_1"))
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
	executor1ID := GenerateRandomID()
	executor2ID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	var attributes []Attribute

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process := CreateProcessFromDB(funcSpec, GenerateRandomID(), GenerateRandomID(), true, RUNNING, time.Time{}, time.Time{}, time.Time{}, time.Time{}, time.Time{}, []string{"errormsg"}, 2, attributes)

	processingTime := int64(process.ProcessingTime())
	assert.True(t, processingTime > 0)

	process.SetState(WAITING)
	processingTime = int64(process.ProcessingTime())
	assert.True(t, processingTime == 0)
}

func TestProcessClone(t *testing.T) {
	colonyID := GenerateRandomID()
	executorType := "test_executor_type"
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3

	funcSpec := CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, maxWaitTime, maxExecTime, maxRetries, make(map[string]string), []string{}, 1, "test_label")
	process := CreateProcess(funcSpec)

	processClone := process.Clone()
	processClone.ID = GenerateRandomID()
	processClone.FunctionSpec.FuncName = "test_func2"

	assert.False(t, processClone.Equals(process))
}
