package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCResolveGeneratorMsg(t *testing.T) {
	msg := CreateResolveGeneratorMsg(core.GenerateRandomID(), core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateResolveGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResolveGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCResolveGeneratorMsgIndent(t *testing.T) {
	msg := CreateResolveGeneratorMsg(core.GenerateRandomID(), core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateResolveGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResolveGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCResolveGeneratorMsgEquals(t *testing.T) {
	msg := CreateResolveGeneratorMsg(core.GenerateRandomID(), core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
