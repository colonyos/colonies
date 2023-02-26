package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowSpecJSON(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	workflowSpec := CreateWorkflowSpec(GenerateRandomID())

	funcSpec1 := CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"

	funcSpec2 := CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"

	funcSpec3 := CreateEmptyFunctionSpec()
	funcSpec3.NodeName = "task3"

	funcSpec4 := CreateEmptyFunctionSpec()
	funcSpec4.NodeName = "task4"

	funcSpec2.AddDependency("task1")
	funcSpec3.AddDependency("task1")
	funcSpec4.AddDependency("task2")
	funcSpec4.AddDependency("task3")

	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	workflowSpec.AddFunctionSpec(funcSpec3)
	workflowSpec.AddFunctionSpec(funcSpec4)

	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)

	workflowSpec2, err := ConvertJSONToWorkflowSpec(jsonStr)
	assert.Nil(t, err)
	assert.True(t, workflowSpec.Equals(workflowSpec2))
}
