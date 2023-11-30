package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveColonyMsg(t *testing.T) {
	msg := CreateRemoveColonyMsg("test_colony_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveColonyMsgIndent(t *testing.T) {
	msg := CreateRemoveColonyMsg("test_colony_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRemoveColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveColonyMsgEquals(t *testing.T) {
	msg := CreateRemoveColonyMsg("test_colony_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
