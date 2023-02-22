package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetFunctionsMsg(t *testing.T) {
	msg := CreateGetFunctionsMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFunctionsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFunctionsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFunctionsMsgIndent(t *testing.T) {
	msg := CreateGetFunctionsMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFunctionsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFunctionsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFunctionsMsgEquals(t *testing.T) {
	msg := CreateGetFunctionsMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
