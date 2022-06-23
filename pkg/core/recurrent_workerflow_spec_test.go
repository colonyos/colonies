package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRecurrentWorkflowSpec(t *testing.T) {
	workflowSpec := CreateWorkflowSpec(GenerateRandomID())
	processSpec1 := CreateEmptyProcessSpec()
	processSpec1.Name = "task1"
	processSpec2 := CreateEmptyProcessSpec()
	processSpec2.Name = "task2"
	processSpec2.AddDependency("task1")
	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	recWorkflow := CreateRecurrentWorkflowSpec("test_recwfname", jsonStr, "*/1 * * * * *") // every second
	jsonStr, err = recWorkflow.ToJSON()
	assert.Nil(t, err)

	recWorkflow2, err := ConvertJSONToRecurrentWorkflowSpec(jsonStr)
	assert.Nil(t, err)
	assert.True(t, recWorkflow.Equals(recWorkflow2))

	workflowSpec2, err := ConvertJSONToWorkflowSpec(recWorkflow2.WorkflowSpec)
	assert.Nil(t, err)
	assert.True(t, workflowSpec.Equals(workflowSpec2))
}

func TestCreateRecurrentWorkflowSpecArray(t *testing.T) {
	var arr []*RecurrentWorkflowSpec
	workflowSpec1 := CreateWorkflowSpec(GenerateRandomID())
	processSpec1 := CreateEmptyProcessSpec()
	processSpec1.Name = "task1"
	processSpec2 := CreateEmptyProcessSpec()
	processSpec2.Name = "task2"
	processSpec2.AddDependency("task1")
	workflowSpec1.AddProcessSpec(processSpec1)
	workflowSpec1.AddProcessSpec(processSpec2)
	jsonStr, err := workflowSpec1.ToJSON()
	assert.Nil(t, err)
	recWorkflow1 := CreateRecurrentWorkflowSpec("test_recwfname1", jsonStr, "*/1 * * * * *") // every second
	arr = append(arr, recWorkflow1)

	recWorkflow2 := CreateRecurrentWorkflowSpec("test_recwfname1", jsonStr, "*/1 * * * * *") // every second
	assert.Nil(t, err)
	arr = append(arr, recWorkflow2)

	jsonStr, err = ConvertRecurrentWorkflowSpecArrayToJSON(arr)
	assert.Nil(t, err)

	_, err = ConvertJSONToRecurrentWorkflowSpecArray(jsonStr + "error")
	assert.NotNil(t, err)

	arr2, err := ConvertJSONToRecurrentWorkflowSpecArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsRecurrentWorkflowSpecArraysEqual(arr, arr2))
	assert.False(t, IsRecurrentWorkflowSpecArraysEqual(arr, nil))
	assert.False(t, IsRecurrentWorkflowSpecArraysEqual(nil, arr2))
	assert.False(t, IsRecurrentWorkflowSpecArraysEqual(nil, nil))
}
