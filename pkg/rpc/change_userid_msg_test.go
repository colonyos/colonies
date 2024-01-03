package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCChangeUserIDMsg(t *testing.T) {
	msg := CreateChangeUserIDMsg("test_colony_name", "test_user_id")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateChangeUserIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeUserIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeUserIDMsgIndent(t *testing.T) {
	msg := CreateChangeUserIDMsg("test_colony_name", "test_user_id")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateChangeUserIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeUserIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeUserIDMsgEquals(t *testing.T) {
	msg := CreateChangeUserIDMsg("test_colony_name", "test_user_id")

	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
