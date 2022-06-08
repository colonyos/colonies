package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCDeleteAllProcessGraphsMsg(t *testing.T) {
	msg := CreateDeleteAllProcessGraphsMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteAllProcessGraphsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteAllProcessGraphsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteAllProcessGraphsMsgIndent(t *testing.T) {
	msg := CreateDeleteAllProcessGraphsMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateDeleteAllProcessGraphsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteAllProcessGraphsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteAllProcessGraphsMsgEquals(t *testing.T) {
	msg := CreateDeleteAllProcessGraphsMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
