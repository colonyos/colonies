package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetProcessHistMsg(t *testing.T) {
	msg := CreateGetProcessHistMsg(core.GenerateRandomID(), core.GenerateRandomID(), 1, 2)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessHistMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessHistMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessHistMsgIndent(t *testing.T) {
	msg := CreateGetProcessHistMsg(core.GenerateRandomID(), core.GenerateRandomID(), 1, 2)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessHistMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessHistMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessHistMsgEquals(t *testing.T) {
	msg := CreateGetProcessHistMsg(core.GenerateRandomID(), core.GenerateRandomID(), 1, 2)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
