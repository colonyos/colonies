package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileMsg(t *testing.T) {
	msg := CreateGetFileMsg("test_colonyid", "test_fileid", "test_prefix", "test_name", false)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFileMsgIndent(t *testing.T) {
	msg := CreateGetFileMsg("test_colonyid", "test_fileid", "test_prefix", "test_name", false)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFileMsgEquals(t *testing.T) {
	msg := CreateGetFileMsg("test_colonyid", "test_fileid", "test_prefix", "test_name", false)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
