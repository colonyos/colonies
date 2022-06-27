package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetGeneratorMsg(t *testing.T) {
	msg := CreateGetGeneratorMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetGeneratorMsgIndent(t *testing.T) {
	msg := CreateGetGeneratorMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetGeneratorMsgEquals(t *testing.T) {
	msg := CreateGetGeneratorMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
