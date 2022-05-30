package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetProcessGraphMsg(t *testing.T) {
	msg := CreateGetProcessGraphMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessGraphMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessGraphMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessGraphMsgIndent(t *testing.T) {
	msg := CreateGetProcessGraphMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessGraphMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessGraphMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessGraphMsgEquals(t *testing.T) {
	msg := CreateGetProcessGraphMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
