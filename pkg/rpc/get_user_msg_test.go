package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetUserMsg(t *testing.T) {
	msg := CreateGetUserMsg("test_colony_name", "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetUserMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetUserMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetUserMsgIndent(t *testing.T) {
	msg := CreateGetUserMsg("test_colony_name", "test_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetUserMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetUserMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetUserMsgEquals(t *testing.T) {
	msg := CreateGetUserMsg("test_colony_name", "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
