package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func createFunction() *core.Function {
	return &core.Function{ExecutorID: core.GenerateRandomID(), ColonyID: core.GenerateRandomID(), FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
}

func TestRPCAddFunctionMsg(t *testing.T) {
	function := createFunction()

	msg := CreateAddFunctionMsg(function)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddFunctionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddFunctionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddFunctionMsgIndent(t *testing.T) {
	function := createFunction()

	msg := CreateAddFunctionMsg(function)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddFunctionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddFunctionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddFunctionMsgEquals(t *testing.T) {
	function := createFunction()

	msg := CreateAddFunctionMsg(function)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
