package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveFileMsg(t *testing.T) {
	msg := CreateRemoveFileMsg("test_colony", "test_fileid", "test_prefix", "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGRemoveFileMsgIndent(t *testing.T) {
	msg := CreateRemoveFileMsg("test_colony", "test_fileid", "test_prefix", "test_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRemoveFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveFileMsgEquals(t *testing.T) {
	msg := CreateRemoveFileMsg("test_colony", "test_fileid", "test_prefix", "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
