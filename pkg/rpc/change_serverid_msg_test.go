package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCChangeServerIDMsg(t *testing.T) {
	msg := CreateChangeServerIDMsg("test_server_id")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateChangeServerIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeServerIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeServerIDMsgIndent(t *testing.T) {
	msg := CreateChangeServerIDMsg("test_server_id")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateChangeServerIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeServerIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeServerIDMsgEquals(t *testing.T) {
	msg := CreateChangeServerIDMsg("test_server_id")

	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
