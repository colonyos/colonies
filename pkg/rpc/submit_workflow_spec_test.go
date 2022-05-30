package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createWorkflowSpec() *core.WorkflowSpec {
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	workflowSpec := core.CreateWorkflowSpec(true)

	processSpec1 := core.CreateEmptyProcessSpec()
	processSpec1.Name = "task1"

	processSpec2 := core.CreateEmptyProcessSpec()
	processSpec2.Name = "task2"

	processSpec3 := core.CreateEmptyProcessSpec()
	processSpec3.Name = "task3"

	processSpec4 := core.CreateEmptyProcessSpec()
	processSpec4.Name = "task4"

	processSpec2.AddDependency("task1")
	processSpec3.AddDependency("task1")
	processSpec4.AddDependency("task2")
	processSpec4.AddDependency("task3")

	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	workflowSpec.AddProcessSpec(processSpec3)
	workflowSpec.AddProcessSpec(processSpec4)

	return workflowSpec
}

func TestRPCSubmitWorkflowSpecMsg(t *testing.T) {
	msg := CreateSubmitWorkflowSpecMsg(createWorkflowSpec())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubmitWorkflowSpecMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubmitWorkflowSpecMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubmitWorkflowSpecMsgIndent(t *testing.T) {
	msg := CreateSubmitWorkflowSpecMsg(createWorkflowSpec())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSubmitWorkflowSpecMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubmitWorkflowSpecMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubmitWorkflowSpecMsgEquals(t *testing.T) {
	msg := CreateSubmitWorkflowSpecMsg(createWorkflowSpec())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
