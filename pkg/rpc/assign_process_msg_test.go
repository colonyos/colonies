package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAssignProcessMsg(t *testing.T) {
	msg := CreateAssignProcessMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAssignProcessMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAssignProcessMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAssignProcessMsgIndent(t *testing.T) {
	msg := CreateAssignProcessMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAssignProcessMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAssignProcessMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAssignProcessMsgEquals(t *testing.T) {
	msg := CreateAssignProcessMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
