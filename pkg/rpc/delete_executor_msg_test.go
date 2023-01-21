package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCDeleteExecutorMsg(t *testing.T) {
	msg := CreateDeleteExecutorMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteExecutorMsgIndent(t *testing.T) {
	msg := CreateDeleteExecutorMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateDeleteExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteExecutorMsgEquals(t *testing.T) {
	msg := CreateDeleteExecutorMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
