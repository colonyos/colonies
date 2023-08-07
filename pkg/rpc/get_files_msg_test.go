package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilesMsg(t *testing.T) {
	msg := CreateGetFilesMsg("test_prefix")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFilesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFilesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFilesMsgIndent(t *testing.T) {
	msg := CreateGetFilesMsg("test_prefix")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFilesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFilesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFilesMsgEquals(t *testing.T) {
	msg := CreateGetFilesMsg("test_prefix")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
