package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetLocationsMsg(t *testing.T) {
	msg := CreateGetLocationsMsg("test_colony")
	assert.Equal(t, GetLocationsPayloadType, msg.MsgType)
	assert.Equal(t, "test_colony", msg.ColonyName)

	jsonStr, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetLocationsMsgFromJSON(jsonStr)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetLocationsMsgToJSONIndent(t *testing.T) {
	msg := CreateGetLocationsMsg("test_colony")

	jsonStr, err := msg.ToJSONIndent()
	assert.Nil(t, err)
	assert.Contains(t, jsonStr, "\n")
}

func TestRPCGetLocationsMsgEqualsNil(t *testing.T) {
	msg := CreateGetLocationsMsg("test_colony")
	assert.False(t, msg.Equals(nil))
}

func TestRPCGetLocationsMsgEqualsDifferentColony(t *testing.T) {
	msg1 := CreateGetLocationsMsg("test_colony1")
	msg2 := CreateGetLocationsMsg("test_colony2")
	assert.False(t, msg1.Equals(msg2))
}

func TestRPCGetLocationsMsgFromJSONInvalid(t *testing.T) {
	_, err := CreateGetLocationsMsgFromJSON("invalid json")
	assert.NotNil(t, err)
}
