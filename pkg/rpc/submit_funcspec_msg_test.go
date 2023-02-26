package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createFunctionSpec() *core.FunctionSpec {
	colonyID := core.GenerateRandomID()
	executorType := "test_executor_type"
	executor1ID := core.GenerateRandomID()
	executor2ID := core.GenerateRandomID()
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3
	env := make(map[string]string)
	env["test_key"] = "test_value"

	return core.CreateFunctionSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, env, []string{}, 1, "test_label")
}

func TestRPCSubmitFunctionSpecMsg(t *testing.T) {
	msg := CreateSubmitFunctionSpecMsg(createFunctionSpec())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubmitFunctionSpecMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubmitFunctionSpecMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubmitFunctionSpecMsgIndent(t *testing.T) {
	msg := CreateSubmitFunctionSpecMsg(createFunctionSpec())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSubmitFunctionSpecMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubmitFunctionSpecMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubmitFunctionSpecMsgEquals(t *testing.T) {
	msg := CreateSubmitFunctionSpecMsg(createFunctionSpec())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
