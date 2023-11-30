package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveAllProcessGraphsMsg(t *testing.T) {
	msg := CreateRemoveAllProcessGraphsMsg(core.GenerateRandomID())
	msg.State = core.NOTSET
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveAllProcessGraphsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveAllProcessGraphsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveAllProcessGraphsMsgIndent(t *testing.T) {
	msg := CreateRemoveAllProcessGraphsMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRemoveAllProcessGraphsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveAllProcessGraphsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveAllProcessGraphsMsgEquals(t *testing.T) {
	msg := CreateRemoveAllProcessGraphsMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
