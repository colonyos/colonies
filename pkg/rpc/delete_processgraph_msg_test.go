package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCDeleteProcessGraphMsg(t *testing.T) {
	msg := CreateDeleteProcessGraphMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteProcessGraphMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteProcessGraphMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteProcessGraphMsgIndent(t *testing.T) {
	msg := CreateDeleteProcessGraphMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateDeleteProcessGraphMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteProcessGraphMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteProcessGraphMsgEquals(t *testing.T) {
	msg := CreateDeleteProcessGraphMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
