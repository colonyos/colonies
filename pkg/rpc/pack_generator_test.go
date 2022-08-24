package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCPackGeneratorMsg(t *testing.T) {
	msg := CreatePackGeneratorMsg(core.GenerateRandomID(), "arg")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreatePackGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreatePackGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCPackGeneratorMsgIndent(t *testing.T) {
	msg := CreatePackGeneratorMsg(core.GenerateRandomID(), "arg")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreatePackGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreatePackGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCPackGeneratorMsgEquals(t *testing.T) {
	msg := CreatePackGeneratorMsg(core.GenerateRandomID(), "arg")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
