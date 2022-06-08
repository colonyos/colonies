package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddAttributeMsg(t *testing.T) {
	key := "test_key"
	value := "test_value"
	attribute := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.OUT, key, value)

	msg := CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddAttributeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddAttributeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddAttributeMsgIndent(t *testing.T) {
	key := "test_key"
	value := "test_value"
	attribute := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), core.GenerateRandomID(), core.OUT, key, value)

	msg := CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddAttributeMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddAttributeMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddAttributeMsgEquals(t *testing.T) {
	key := "test_key"
	value := "test_value"
	attribute := core.CreateAttribute(core.GenerateRandomID(), core.GenerateRandomID(), "", core.OUT, key, value)

	msg := CreateAddAttributeMsg(attribute)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
