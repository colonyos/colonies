package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetLocationMsg(t *testing.T) {
	msg := CreateGetLocationMsg("test_colony", "test_name")
	assert.Equal(t, GetLocationPayloadType, msg.MsgType)
	assert.Equal(t, "test_colony", msg.ColonyName)
	assert.Equal(t, "test_name", msg.Name)

	jsonStr, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetLocationMsgFromJSON(jsonStr)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetLocationMsgToJSONIndent(t *testing.T) {
	msg := CreateGetLocationMsg("test_colony", "test_name")

	jsonStr, err := msg.ToJSONIndent()
	assert.Nil(t, err)
	assert.Contains(t, jsonStr, "\n")
}

func TestRPCGetLocationMsgEqualsNil(t *testing.T) {
	msg := CreateGetLocationMsg("test_colony", "test_name")
	assert.False(t, msg.Equals(nil))
}

func TestRPCGetLocationMsgEqualsDifferentColony(t *testing.T) {
	msg1 := CreateGetLocationMsg("test_colony1", "test_name")
	msg2 := CreateGetLocationMsg("test_colony2", "test_name")
	assert.False(t, msg1.Equals(msg2))
}

func TestRPCGetLocationMsgEqualsDifferentName(t *testing.T) {
	msg1 := CreateGetLocationMsg("test_colony", "test_name1")
	msg2 := CreateGetLocationMsg("test_colony", "test_name2")
	assert.False(t, msg1.Equals(msg2))
}

func TestRPCGetLocationMsgFromJSONInvalid(t *testing.T) {
	_, err := CreateGetLocationMsgFromJSON("invalid json")
	assert.NotNil(t, err)
}
