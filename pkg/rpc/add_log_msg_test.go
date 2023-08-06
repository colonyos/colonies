package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddLogMsg(t *testing.T) {
	msg := CreateAddLogMsg(core.GenerateRandomID(), "test_msg")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddLogMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddLogMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddLogMsgIndent(t *testing.T) {
	msg := CreateAddLogMsg(core.GenerateRandomID(), "test_msg")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddLogMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddLogMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddLogMsgEquals(t *testing.T) {
	msg := CreateAddLogMsg(core.GenerateRandomID(), "test_msg")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
