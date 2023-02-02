package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCApproveExecutorMsg(t *testing.T) {
	msg := CreateApproveExecutorMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateApproveExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateApproveExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCApproveExecutorMsgIndent(t *testing.T) {
	msg := CreateApproveExecutorMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateApproveExecutorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateApproveExecutorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCApproveExecutorMsgEquals(t *testing.T) {
	msg := CreateApproveExecutorMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
