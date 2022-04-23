package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetAttributeMsg(t *testing.T) {
	msg := CreateGetAttributeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetAttributeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetAttributeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetAttributeMsgIndent(t *testing.T) {
	msg := CreateGetAttributeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetAttributeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetAttributeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetAttributeMsgEquals(t *testing.T) {
	msg := CreateGetAttributeMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
