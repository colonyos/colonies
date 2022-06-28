package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetGeneratorsMsg(t *testing.T) {
	msg := CreateGetGeneratorsMsg(core.GenerateRandomID(), 2)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetGeneratorsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetGeneratorsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetGeneratorsMsgIndent(t *testing.T) {
	msg := CreateGetGeneratorsMsg(core.GenerateRandomID(), 2)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetGeneratorsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetGeneratorsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetGeneratorsMsgEquals(t *testing.T) {
	msg := CreateGetGeneratorsMsg(core.GenerateRandomID(), 2)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
