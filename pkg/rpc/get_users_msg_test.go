package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetUsersMsg(t *testing.T) {
	msg := CreateGetUsersMsg("test_colony_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetUsersMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetUsersMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetUsersMsgIndent(t *testing.T) {
	msg := CreateGetUsersMsg("test_colony_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetUsersMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetUsersMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetUsersMsgEquals(t *testing.T) {
	msg := CreateGetUsersMsg("test_colony_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
