package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetNodesByLocationMsg(t *testing.T) {
	msg := CreateGetNodesByLocationMsg("test_colony_name", "test_location")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetNodesByLocationMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetNodesByLocationMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetNodesByLocationMsgIndent(t *testing.T) {
	msg := CreateGetNodesByLocationMsg("test_colony_name", "test_location")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetNodesByLocationMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetNodesByLocationMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetNodesByLocationMsgEquals(t *testing.T) {
	msg := CreateGetNodesByLocationMsg("test_colony_name", "test_location")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
