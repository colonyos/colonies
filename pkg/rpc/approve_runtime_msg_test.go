package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCApproveRuntimeMsg(t *testing.T) {
	msg := CreateApproveRuntimeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateApproveRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateApproveRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCApproveRuntimeMsgIndent(t *testing.T) {
	msg := CreateApproveRuntimeMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateApproveRuntimeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateApproveRuntimeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}
