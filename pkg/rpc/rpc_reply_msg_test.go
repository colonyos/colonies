package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCReplyMsg(t *testing.T) {
	msg, err := CreateRPCReplyMsg("test_method", "test_payload")
	assert.Nil(t, err)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRPCReplyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRPCReplyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))

	assert.Equal(t, msg.DecodePayload(), "test_payload")
}

func TestRPCReplyMsgIndent(t *testing.T) {
	msg, err := CreateRPCReplyMsg("test_method", "test_payload")
	assert.Nil(t, err)

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRPCReplyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRPCReplyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))

	assert.Equal(t, msg.DecodePayload(), "test_payload")
}

func TestRPCReplyMsgEquals(t *testing.T) {
	msg, err := CreateRPCReplyMsg("test_method", "test_payload")
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
