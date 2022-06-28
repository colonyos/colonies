package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddGeneratorMsg(t *testing.T) {
	generator := utils.FakeGenerator(t, core.GenerateRandomID())
	msg := CreateAddGeneratorMsg(generator)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddGeneratorMsgIndent(t *testing.T) {
	generator := utils.FakeGenerator(t, core.GenerateRandomID())
	msg := CreateAddGeneratorMsg(generator)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddGeneratorMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddGeneratorMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddGeneratorMsgEquals(t *testing.T) {
	generator := utils.FakeGenerator(t, core.GenerateRandomID())
	msg := CreateAddGeneratorMsg(generator)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
