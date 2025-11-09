package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetNodeMsg(t *testing.T) {
	msg := CreateGetNodeMsg("test_colony_name", "test_node_name")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetNodeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetNodeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetNodeMsgIndent(t *testing.T) {
	msg := CreateGetNodeMsg("test_colony_name", "test_node_name")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetNodeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetNodeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetNodeMsgEquals(t *testing.T) {
	msg := CreateGetNodeMsg("test_colony_name", "test_node_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
