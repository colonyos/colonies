package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetNodesMsg(t *testing.T) {
	msg := CreateGetNodesMsg("test_colony_name")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetNodesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetNodesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetNodesMsgIndent(t *testing.T) {
	msg := CreateGetNodesMsg("test_colony_name")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetNodesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetNodesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetNodesMsgEquals(t *testing.T) {
	msg := CreateGetNodesMsg("test_colony_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
