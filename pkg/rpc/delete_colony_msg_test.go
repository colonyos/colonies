package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCDeleteColonyMsg(t *testing.T) {
	msg := CreateDeleteColonyMsg("test_colony_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteColonyMsgIndent(t *testing.T) {
	msg := CreateDeleteColonyMsg("test_colony_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateDeleteColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteColonyMsgEquals(t *testing.T) {
	msg := CreateDeleteColonyMsg("test_colony_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
