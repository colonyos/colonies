package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCRenameColonyMsg(t *testing.T) {
	msg := CreateRenameColonyMsg(core.GenerateRandomID(), "new_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRenameColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRenameColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDRenameColonyMsgIndent(t *testing.T) {
	msg := CreateRenameColonyMsg(core.GenerateRandomID(), "new_name")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRenameColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRenameColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRenameColonyMsgEquals(t *testing.T) {
	msg := CreateRenameColonyMsg(core.GenerateRandomID(), "new_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
