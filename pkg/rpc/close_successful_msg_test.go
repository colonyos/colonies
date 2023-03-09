package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCCloseSuccessfulMsg(t *testing.T) {
	msg := CreateCloseSuccessfulMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateCloseSuccessfulMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateCloseSuccessfulMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCCloseSuccessfulMsgWithResult(t *testing.T) {
	msg := CreateCloseSuccessfulMsg(core.GenerateRandomID())
	msg.Output = make([]interface{}, 2)
	msg.Output[0] = "result1"
	msg.Output[1] = "result2"
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateCloseSuccessfulMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateCloseSuccessfulMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCCloseSuccessfulMsgIndent(t *testing.T) {
	msg := CreateCloseSuccessfulMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateCloseSuccessfulMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateCloseSuccessfulMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCCloseSuccessfulMsgEquals(t *testing.T) {
	msg := CreateCloseSuccessfulMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
