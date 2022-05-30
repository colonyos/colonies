package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetProcessGraphsMsg(t *testing.T) {
	msg := CreateGetProcessGraphsMsg(core.GenerateRandomID(), 1, 2)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessGraphsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessGraphsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessGraphsMsgIndent(t *testing.T) {
	msg := CreateGetProcessGraphsMsg(core.GenerateRandomID(), 1, 2)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessGraphsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessGraphsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessGraphsMsgEquals(t *testing.T) {
	msg := CreateGetProcessGraphsMsg(core.GenerateRandomID(), 1, 2)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
