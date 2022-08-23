package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGenerator(t *testing.T) {
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
	generator := CreateGenerator(GenerateRandomID(), "test_genname", jsonStr, 10)
	generator.ID = GenerateRandomID()
	jsonStr, err = generator.ToJSON()
	assert.Nil(t, err)

	generator2, err := ConvertJSONToGenerator(jsonStr)
	assert.Nil(t, err)
	assert.True(t, generator.Equals(generator2))

	workflowSpec2, err := ConvertJSONToWorkflowSpec(generator2.WorkflowSpec)
	assert.Nil(t, err)
	assert.True(t, workflowSpec.Equals(workflowSpec2))
}

func TestCreateGeneratorSpecArray(t *testing.T) {
	var arr []*Generator
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
	generator1 := CreateGenerator(GenerateRandomID(), "test_genname1", jsonStr, 10)
	generator1.ID = GenerateRandomID()
	arr = append(arr, generator1)

	generator2 := CreateGenerator(GenerateRandomID(), "test_genname2", jsonStr, 10)
	generator2.ID = GenerateRandomID()
	assert.Nil(t, err)
	arr = append(arr, generator2)

	jsonStr, err = ConvertGeneratorArrayToJSON(arr)
	assert.Nil(t, err)

	_, err = ConvertJSONToGeneratorArray(jsonStr + "error")
	assert.NotNil(t, err)

	arr2, err := ConvertJSONToGeneratorArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsGeneratorArraysEqual(arr, arr2))
	assert.False(t, IsGeneratorArraysEqual(arr, nil))
	assert.False(t, IsGeneratorArraysEqual(nil, arr2))
	assert.False(t, IsGeneratorArraysEqual(nil, nil))
}
