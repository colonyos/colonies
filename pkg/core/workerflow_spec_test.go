package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkflowSpecJSON(t *testing.T) {
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	workflowSpec := CreateWorkflowSpec(GenerateRandomID(), true)

	processSpec1 := CreateEmptyProcessSpec()
	processSpec1.Name = "task1"

	processSpec2 := CreateEmptyProcessSpec()
	processSpec2.Name = "task2"

	processSpec3 := CreateEmptyProcessSpec()
	processSpec3.Name = "task3"

	processSpec4 := CreateEmptyProcessSpec()
	processSpec4.Name = "task4"

	processSpec2.AddDependency("task1")
	processSpec3.AddDependency("task1")
	processSpec4.AddDependency("task2")
	processSpec4.AddDependency("task3")

	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	workflowSpec.AddProcessSpec(processSpec3)
	workflowSpec.AddProcessSpec(processSpec4)

	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)

	fmt.Println(jsonStr)

	workflowSpec2, err := ConvertJSONToWorkflowSpec(jsonStr)
	assert.Nil(t, err)
	assert.True(t, workflowSpec.Equals(workflowSpec2))
}
