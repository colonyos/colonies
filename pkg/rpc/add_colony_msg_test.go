package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddColonyMsg(t *testing.T) {
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	msg := CreateAddColonyMsg(colony)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddColonyMsgIndent(t *testing.T) {
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")
	msg := CreateAddColonyMsg(colony)

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddColonyMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddColonyMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddColonyMsgEquals(t *testing.T) {
	colony := core.CreateColony(core.GenerateRandomID(), "test_colony_name_1")

	msg := CreateAddColonyMsg(colony)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
