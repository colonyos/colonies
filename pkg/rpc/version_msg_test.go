package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCVersionMsg(t *testing.T) {
	msg := CreateVersionMsg("build_version", "build_type")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateVersionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateVersionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCVersionMsgIndent(t *testing.T) {
	msg := CreateVersionMsg("build_version", "build_type")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateVersionMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateVersionMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCVersionMsgEquals(t *testing.T) {
	msg := CreateVersionMsg("build_version", "build_type")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
