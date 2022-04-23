package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetProcessStatMsg(t *testing.T) {
	msg := CreateGetProcessStatMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessStatMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessStatMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessStatMsgIndent(t *testing.T) {
	msg := CreateGetProcessStatMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessStatMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessStatMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessStatMsgEquals(t *testing.T) {
	msg := CreateGetProcessStatMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
