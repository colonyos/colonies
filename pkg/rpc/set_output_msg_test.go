package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCSetOutputMsg(t *testing.T) {
	output := make([]interface{}, 2)
	output[0] = "result1"
	output[1] = "result2"
	msg := CreateSetOutputMsg(core.GenerateRandomID(), output)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSetOutputMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSetOutputMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSetOutputMsgIndent(t *testing.T) {
	output := make([]interface{}, 1)
	output[0] = "result1"
	msg := CreateSetOutputMsg(core.GenerateRandomID(), output)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSetOutputMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSetOutputMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSetOutputMsgEquals(t *testing.T) {
	output := make([]interface{}, 1)
	output[0] = "result1"
	msg := CreateSetOutputMsg(core.GenerateRandomID(), output)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
