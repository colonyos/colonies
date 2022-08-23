package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddArgGeneratorMsg(t *testing.T) {
	msg := CreateAddArgGeneratorMsg(core.GenerateRandomID(), "arg")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddArgGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddArgGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddArgGeneratorMsgIndent(t *testing.T) {
	msg := CreateAddArgGeneratorMsg(core.GenerateRandomID(), "arg")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddArgGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddArgGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddArgGeneratorMsgEquals(t *testing.T) {
	msg := CreateAddArgGeneratorMsg(core.GenerateRandomID(), "arg")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
