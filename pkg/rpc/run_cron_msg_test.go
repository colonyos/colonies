package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRunCronMsg(t *testing.T) {
	msg := CreateRunCronMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRunCronMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRunCronMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRunCronMsgIndent(t *testing.T) {
	msg := CreateRunCronMsg(core.GenerateRandomID())
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRunCronMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRunCronMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRunCronMsgEquals(t *testing.T) {
	msg := CreateRunCronMsg(core.GenerateRandomID())
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
