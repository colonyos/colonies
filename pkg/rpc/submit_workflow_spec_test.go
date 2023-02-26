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

	workflowSpec := core.CreateWorkflowSpec(core.GenerateRandomID())

	funcSpec1 := core.CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"

	funcSpec2 := core.CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"

	funcSpec3 := core.CreateEmptyFunctionSpec()
	funcSpec3.NodeName = "task3"

	funcSpec4 := core.CreateEmptyFunctionSpec()
	funcSpec4.NodeName = "task4"

	funcSpec2.AddDependency("task1")
	funcSpec3.AddDependency("task1")
	funcSpec4.AddDependency("task2")
	funcSpec4.AddDependency("task3")

	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	workflowSpec.AddFunctionSpec(funcSpec3)
	workflowSpec.AddFunctionSpec(funcSpec4)

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
