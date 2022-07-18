package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetClusterMsg(t *testing.T) {
	msg := CreateGetClusterMsg()
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetClusterMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetClusterMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetClusterMsgIndent(t *testing.T) {
	msg := CreateGetClusterMsg()
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetClusterMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetClusterMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetClusterMsgEquals(t *testing.T) {
	msg := CreateGetClusterMsg()
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
