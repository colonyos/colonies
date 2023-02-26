package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateGenerator(t *testing.T) {
	workflowSpec := CreateWorkflowSpec(GenerateRandomID())
	funcSpec1 := CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"
	funcSpec2 := CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"
	funcSpec2.AddDependency("task1")
	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := CreateGenerator(GenerateRandomID(), "test_genname", jsonStr, 10)
	generator.ID = GenerateRandomID()
	generator.QueueSize = 100
	generator.CheckerPeriod = 200
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
	funcSpec1 := CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"
	funcSpec2 := CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"
	funcSpec2.AddDependency("task1")
	workflowSpec1.AddFunctionSpec(funcSpec1)
	workflowSpec1.AddFunctionSpec(funcSpec2)
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
