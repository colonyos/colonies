package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetColonyMsg(t *testing.T) {
	msg := CreateGetColonyMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetColonyMsgIndent(t *testing.T) {
	msg := CreateGetColonyMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}
