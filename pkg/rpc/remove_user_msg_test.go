package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveUserMsg(t *testing.T) {
	msg := CreateRemoveUserMsg(core.GenerateRandomID(), "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveUserMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveUserMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveUserMsgIndent(t *testing.T) {
	msg := CreateRemoveUserMsg(core.GenerateRandomID(), "test_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRemoveUserMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveUserMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveUserMsgEquals(t *testing.T) {
	msg := CreateRemoveUserMsg(core.GenerateRandomID(), "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
