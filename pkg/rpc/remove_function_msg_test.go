package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveFunctionMsg(t *testing.T) {
	msg := CreateRemoveFunctionMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveFunctionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveFunctionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveFunctionMsgIndent(t *testing.T) {
	msg := CreateRemoveFunctionMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRemoveFunctionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveFunctionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveFunctionMsgEquals(t *testing.T) {
	msg := CreateRemoveFunctionMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
