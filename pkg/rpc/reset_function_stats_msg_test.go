package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCResetFunctionStatsMsg(t *testing.T) {
	msg := CreateResetFunctionStatsMsg("test-colony")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateResetFunctionStatsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResetFunctionStatsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCResetFunctionStatsMsgIndent(t *testing.T) {
	msg := CreateResetFunctionStatsMsg("test-colony")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateResetFunctionStatsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateResetFunctionStatsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCResetFunctionStatsMsgEquals(t *testing.T) {
	msg := CreateResetFunctionStatsMsg("test-colony")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
