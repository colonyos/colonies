package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFileMsg(t *testing.T) {
	msg := CreateDeleteFileMsg("test_colony", "test_fileid", "test_prefix", "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGDeleteFileMsgIndent(t *testing.T) {
	msg := CreateDeleteFileMsg("test_colony", "test_fileid", "test_prefix", "test_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateDeleteFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteFileMsgEquals(t *testing.T) {
	msg := CreateDeleteFileMsg("test_colony", "test_fileid", "test_prefix", "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
