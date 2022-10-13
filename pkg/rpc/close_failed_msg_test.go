package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCCloseFailedMsg(t *testing.T) {
	msg := CreateCloseFailedMsg(core.GenerateRandomID(), []string{"error"})
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateCloseFailedMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateCloseFailedMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCCloseFailedMsgIndent(t *testing.T) {
	msg := CreateCloseFailedMsg(core.GenerateRandomID(), []string{"error"})
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateCloseFailedMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateCloseFailedMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCCloseFailedMsgEquals(t *testing.T) {
	msg := CreateCloseFailedMsg(core.GenerateRandomID(), []string{"error"})
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
