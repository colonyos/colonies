package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCResetDatabaseMsg(t *testing.T) {
	msg := CreateResetDatabaseMsg()
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateResetDatabaseMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResetDatabaseMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCResetDatabaseMsgIndent(t *testing.T) {
	msg := CreateResetDatabaseMsg()
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateResetDatabaseMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResetDatabaseMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCResetDatabaseMsgEquals(t *testing.T) {
	msg := CreateResetDatabaseMsg()
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
