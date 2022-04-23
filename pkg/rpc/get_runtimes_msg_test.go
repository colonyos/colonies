package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetRuntimeMsg(t *testing.T) {
	msg := CreateGetRuntimeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetRuntimeMsgIndent(t *testing.T) {
	msg := CreateGetRuntimeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetRuntimeMsgEquals(t *testing.T) {
	msg := CreateGetRuntimeMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
