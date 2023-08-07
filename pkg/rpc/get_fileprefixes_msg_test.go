package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFilePrefixesMsg(t *testing.T) {
	msg := CreateGetFilePrefixesMsg()
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFilePrefixesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFilePrefixesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFilePrefixesMsgIndent(t *testing.T) {
	msg := CreateGetFilePrefixesMsg()
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFilePrefixesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFilePrefixesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFilePrefixesMsgEquals(t *testing.T) {
	msg := CreateGetFilePrefixesMsg()
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
