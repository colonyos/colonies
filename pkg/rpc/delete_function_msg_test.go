package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCDeleteFunctionMsg(t *testing.T) {
	msg := CreateDeleteFunctionMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteFunctionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteFunctionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteFunctionMsgIndent(t *testing.T) {
	msg := CreateDeleteFunctionMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateDeleteFunctionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteFunctionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteFunctionMsgEquals(t *testing.T) {
	msg := CreateDeleteFunctionMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
