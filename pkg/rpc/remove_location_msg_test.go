package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveLocationMsg(t *testing.T) {
	msg := CreateRemoveLocationMsg("test_colony", "test_name")
	assert.Equal(t, RemoveLocationPayloadType, msg.MsgType)
	assert.Equal(t, "test_colony", msg.ColonyName)
	assert.Equal(t, "test_name", msg.Name)

	jsonStr, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveLocationMsgFromJSON(jsonStr)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveLocationMsgToJSONIndent(t *testing.T) {
	msg := CreateRemoveLocationMsg("test_colony", "test_name")

	jsonStr, err := msg.ToJSONIndent()
	assert.Nil(t, err)
	assert.Contains(t, jsonStr, "\n")
}

func TestRPCRemoveLocationMsgEqualsNil(t *testing.T) {
	msg := CreateRemoveLocationMsg("test_colony", "test_name")
	assert.False(t, msg.Equals(nil))
}

func TestRPCRemoveLocationMsgEqualsDifferentColony(t *testing.T) {
	msg1 := CreateRemoveLocationMsg("test_colony1", "test_name")
	msg2 := CreateRemoveLocationMsg("test_colony2", "test_name")
	assert.False(t, msg1.Equals(msg2))
}

func TestRPCRemoveLocationMsgEqualsDifferentName(t *testing.T) {
	msg1 := CreateRemoveLocationMsg("test_colony", "test_name1")
	msg2 := CreateRemoveLocationMsg("test_colony", "test_name2")
	assert.False(t, msg1.Equals(msg2))
}

func TestRPCRemoveLocationMsgFromJSONInvalid(t *testing.T) {
	_, err := CreateRemoveLocationMsgFromJSON("invalid json")
	assert.NotNil(t, err)
}
