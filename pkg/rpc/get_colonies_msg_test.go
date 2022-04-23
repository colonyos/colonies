package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetColoniesMsg(t *testing.T) {
	msg := CreateGetColoniesMsg()
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetColoniesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetColoniesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetColoniesMsgIndent(t *testing.T) {
	msg := CreateGetColoniesMsg()
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetColoniesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetColoniesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetColoniesMsgEquals(t *testing.T) {
	msg := CreateGetColoniesMsg()
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
