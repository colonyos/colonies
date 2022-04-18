package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCSubscribeProcessMsg(t *testing.T) {
	msg := CreateSubscribeProcessMsg(core.GenerateRandomID(), 1, 2)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeProcessMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubscribeProcessMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubscribeProcessMsgIndent(t *testing.T) {
	msg := CreateSubscribeProcessMsg(core.GenerateRandomID(), 1, 2)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeProcessMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubscribeProcessMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}
