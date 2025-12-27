package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddLocationMsg(t *testing.T) {
	location := core.CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	msg := CreateAddLocationMsg(location)
	assert.Equal(t, AddLocationPayloadType, msg.MsgType)
	assert.True(t, location.Equals(msg.Location))

	jsonStr, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddLocationMsgFromJSON(jsonStr)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddLocationMsgToJSONIndent(t *testing.T) {
	location := core.CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	msg := CreateAddLocationMsg(location)

	jsonStr, err := msg.ToJSONIndent()
	assert.Nil(t, err)
	assert.Contains(t, jsonStr, "\n")
}

func TestRPCAddLocationMsgEqualsNil(t *testing.T) {
	location := core.CreateLocation("test_id", "test_name", "test_colony", "test_description", 12.34, 56.78)
	msg := CreateAddLocationMsg(location)
	assert.False(t, msg.Equals(nil))
}

func TestRPCAddLocationMsgFromJSONInvalid(t *testing.T) {
	_, err := CreateAddLocationMsgFromJSON("invalid json")
	assert.NotNil(t, err)
}
