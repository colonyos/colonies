package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateProcess(t *testing.T) {
	colonyID := GenerateRandomID()
	computer1ID := GenerateRandomID()
	computer2ID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	process := CreateProcess(colonyID, []string{computer1ID, computer2ID}, computerType, timeout, maxRetries, mem, cores, gpus)

	assert.Equal(t, colonyID, process.TargetColonyID)
	assert.Contains(t, process.TargetComputerIDs, computer1ID)
	assert.Contains(t, process.TargetComputerIDs, computer2ID)
	assert.Equal(t, computerType, process.ComputerType)
	assert.Equal(t, timeout, process.Timeout)
	assert.Equal(t, maxRetries, process.MaxRetries)
	assert.Equal(t, mem, process.Mem)
	assert.Equal(t, cores, process.Cores)
	assert.Equal(t, gpus, process.GPUs)
	assert.False(t, process.IsAssigned)
	process.Assign()
	assert.True(t, process.IsAssigned)
	process.Unassign()
	assert.False(t, process.IsAssigned)
}

func TestTimeCalc(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	process := CreateProcess(colonyID, []string{}, computerType, timeout, maxRetries, mem, cores, gpus)
	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))
	assert.False(t, process.WaitingTime() < 900000000 && process.WaitingTime() > 1200000000)
	assert.False(t, process.WaitingTime() < 3000000000 && process.WaitingTime() > 4000000000)
}

func TestProcessToJSON(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	process := CreateProcess(colonyID, []string{}, computerType, timeout, maxRetries, mem, cores, gpus)
	process.SetSubmissionTime(startTime)
	process.SetStartTime(startTime.Add(1 * time.Second))
	process.SetEndTime(startTime.Add(4 * time.Second))
	attribute1ID := GenerateRandomID()
	attribute2ID := GenerateRandomID()
	attribute3ID := GenerateRandomID()
	attribute4ID := GenerateRandomID()
	attribute5ID := GenerateRandomID()
	attribute6ID := GenerateRandomID()
	var inAttributes []*Attribute
	inAttributes = append(inAttributes, CreateAttribute(attribute1ID, IN, "in_key_1", "in_value_1"))
	inAttributes = append(inAttributes, CreateAttribute(attribute2ID, IN, "in_key_2", "in_value_2"))
	var errAttributes []*Attribute
	errAttributes = append(errAttributes, CreateAttribute(attribute3ID, ERR, "err_key_1", "err_value_1"))
	errAttributes = append(errAttributes, CreateAttribute(attribute4ID, ERR, "err_key_2", "err_value_2"))
	var outAttributes []*Attribute
	outAttributes = append(outAttributes, CreateAttribute(attribute5ID, OUT, "out_key_1", "out_value_1"))
	outAttributes = append(outAttributes, CreateAttribute(attribute6ID, OUT, "out_key_2", "out_value_2"))
	process.SetInAttributes(inAttributes)
	process.SetErrAttributes(errAttributes)
	process.SetOutAttributes(outAttributes)

	jsonString, err := process.ToJSON()
	assert.Nil(t, err)

	process2, err := ConvertJSONToProcess(jsonString)
	assert.Nil(t, err)

	counter := 0
	for _, attribute := range process2.InAttributes {
		if attribute.TargetID == attribute1ID &&
			attribute.AttributeType == IN &&
			attribute.Key == "in_key_1" &&
			attribute.Value == "in_value_1" {
			counter++
		}
		if attribute.TargetID == attribute2ID &&
			attribute.AttributeType == IN &&
			attribute.Key == "in_key_2" &&
			attribute.Value == "in_value_2" {
			counter++
		}
	}
	assert.Equal(t, 2, counter)

	counter = 0
	for _, attribute := range process2.ErrAttributes {
		if attribute.TargetID == attribute3ID &&
			attribute.AttributeType == ERR &&
			attribute.Key == "err_key_1" &&
			attribute.Value == "err_value_1" {
			counter++
		}
		if attribute.TargetID == attribute4ID &&
			attribute.AttributeType == ERR &&
			attribute.Key == "err_key_2" &&
			attribute.Value == "err_value_2" {
			counter++
		}
	}
	assert.Equal(t, 2, counter)

	counter = 0
	for _, attribute := range process.OutAttributes {
		if attribute.TargetID == attribute5ID &&
			attribute.AttributeType == OUT &&
			attribute.Key == "out_key_1" &&
			attribute.Value == "out_value_1" {
			counter++
		}
		if attribute.TargetID == attribute6ID &&
			attribute.AttributeType == OUT &&
			attribute.Key == "out_key_2" &&
			attribute.Value == "out_value_2" {
			counter++
		}
	}
	assert.Equal(t, 2, counter)

	// TODO: equal
	assert.Equal(t, process.ID, process2.ID)
	assert.Equal(t, process.TargetColonyID, process2.TargetColonyID)
	assert.Equal(t, process.TargetComputerIDs, process2.TargetComputerIDs)
	assert.Equal(t, process.AssignedComputerID, process2.AssignedComputerID)
	assert.Equal(t, process.Status, process2.Status)
	assert.Equal(t, process.IsAssigned, process2.IsAssigned)
	assert.Equal(t, process.ComputerType, process2.ComputerType)
	assert.Equal(t, process.Deadline, process2.Deadline)
	assert.Equal(t, process.Timeout, process2.Timeout)
	assert.Equal(t, process.Retries, process2.Retries)
	assert.Equal(t, process.MaxRetries, process2.MaxRetries)
	assert.Equal(t, process.Mem, process2.Mem)
	assert.Equal(t, process.Cores, process2.Cores)
	assert.Equal(t, process.GPUs, process2.GPUs)
}

func TestProcessArrayToJSON(t *testing.T) {
	startTime := time.Now()

	colonyID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1

	process1 := CreateProcess(colonyID, []string{}, computerType, timeout, maxRetries, mem, cores, gpus)
	process1.SetSubmissionTime(startTime)
	process1.SetStartTime(startTime.Add(1 * time.Second))
	process1.SetEndTime(startTime.Add(4 * time.Second))
	attribute1ID := GenerateRandomID()
	attribute2ID := GenerateRandomID()
	attribute3ID := GenerateRandomID()
	var inAttributes1 []*Attribute
	inAttributes1 = append(inAttributes1, CreateAttribute(attribute1ID, IN, "in_key_1", "in_value_1"))
	var errAttributes1 []*Attribute
	errAttributes1 = append(errAttributes1, CreateAttribute(attribute2ID, ERR, "err_key_1", "err_value_1"))
	var outAttributes1 []*Attribute
	outAttributes1 = append(outAttributes1, CreateAttribute(attribute3ID, OUT, "out_key_1", "out_value_1"))
	process1.SetInAttributes(inAttributes1)
	process1.SetErrAttributes(errAttributes1)
	process1.SetOutAttributes(outAttributes1)

	process2 := CreateProcess(colonyID, []string{}, computerType, timeout, maxRetries, mem, cores, gpus)
	process2.SetSubmissionTime(startTime)
	process2.SetStartTime(startTime.Add(1 * time.Second))
	process2.SetEndTime(startTime.Add(4 * time.Second))
	attribute4ID := GenerateRandomID()
	attribute5ID := GenerateRandomID()
	attribute6ID := GenerateRandomID()
	var inAttributes2 []*Attribute
	inAttributes2 = append(inAttributes2, CreateAttribute(attribute4ID, IN, "in_key_1", "in_value_1"))
	var errAttributes2 []*Attribute
	errAttributes2 = append(errAttributes2, CreateAttribute(attribute5ID, ERR, "err_key_1", "err_value_1"))
	var outAttributes2 []*Attribute
	outAttributes2 = append(outAttributes2, CreateAttribute(attribute6ID, OUT, "out_key_1", "out_value_1"))
	process2.SetInAttributes(inAttributes2)
	process2.SetErrAttributes(errAttributes2)
	process2.SetOutAttributes(outAttributes2)

	var processes []*Process
	processes = append(processes, process1)
	processes = append(processes, process2)

	jsonString, err := ConvertProcessArrayToJSON(processes)
	assert.Nil(t, err)
	processes2, err := ConvertJSONToProcessArray(jsonString)
	assert.Nil(t, err)

	counter := 0
	for _, process := range processes2 {
		if process.ID == process1.ID {
			counter++
		}
		if process.ID == process2.ID {
			counter++
		}
	}

	assert.Equal(t, 2, counter)
}
