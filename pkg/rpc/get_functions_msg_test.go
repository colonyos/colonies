package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetFunctionsByExecutorIDMsg(t *testing.T) {
	msg := CreateGetFunctionsByExecutorIDMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFunctionsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFunctionsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFunctionsByExecutorIDMsgIndent(t *testing.T) {
	msg := CreateGetFunctionsByExecutorIDMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFunctionsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFunctionsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFunctionsByExecutorIDMsgEquals(t *testing.T) {
	msg := CreateGetFunctionsByExecutorIDMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}

func TestRPCGetFunctionsByColonyIDMsg(t *testing.T) {
	msg := CreateGetFunctionsByColonyIDMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFunctionsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFunctionsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFunctionsByColonyIDMsgIndent(t *testing.T) {
	msg := CreateGetFunctionsByColonyIDMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFunctionsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFunctionsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFunctionsByColonyIDMsgEquals(t *testing.T) {
	msg := CreateGetFunctionsByColonyIDMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
