package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGetCronMsg(t *testing.T) {
	msg := CreateGetCronMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetCronMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetCronMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetCronMsgIndent(t *testing.T) {
	msg := CreateGetCronMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetCronMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetCronMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetCronMsgEquals(t *testing.T) {
	msg := CreateGetCronMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
