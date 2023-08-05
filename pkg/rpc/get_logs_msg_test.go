package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestGetLogsMsg(t *testing.T) {
	msg := CreateGetLogsMsg(core.GenerateRandomID(), 100)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetLogsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetLogsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetLogsMsgIndent(t *testing.T) {
	msg := CreateGetLogsMsg(core.GenerateRandomID(), 100)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetLogsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetLogsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetLogsMsgEquals(t *testing.T) {
	msg := CreateGetLogsMsg(core.GenerateRandomID(), 100)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
