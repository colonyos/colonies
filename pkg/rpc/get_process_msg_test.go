package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetProcessMsg(t *testing.T) {
	msg := CreateGetProcessMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetProcessMsgIndent(t *testing.T) {
	msg := CreateGetProcessMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetProcessMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetProcessMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}
