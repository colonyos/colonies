package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCRejctExecutorMsg(t *testing.T) {
	msg := CreateRejectExecutorMsg(core.GenerateRandomID(), core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRejectExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRejectExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRejctExecutorMsgIndent(t *testing.T) {
	msg := CreateRejectExecutorMsg(core.GenerateRandomID(), core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRejectExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRejectExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRejectExecutorMsgEquals(t *testing.T) {
	msg := CreateRejectExecutorMsg(core.GenerateRandomID(), core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
