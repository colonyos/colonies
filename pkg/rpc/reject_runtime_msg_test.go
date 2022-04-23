package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCRejctRuntimeMsg(t *testing.T) {
	msg := CreateRejectRuntimeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRejectRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRejectRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRejctRuntimeMsgIndent(t *testing.T) {
	msg := CreateRejectRuntimeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRejectRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRejectRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRejctRuntimeMsgEquals(t *testing.T) {
	msg := CreateRejectRuntimeMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
