package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRPCGetCronsMsg(t *testing.T) {
	msg := CreateGetCronsMsg(core.GenerateRandomID(), 2)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetCronsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetCronsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetCronsMsgIndent(t *testing.T) {
	msg := CreateGetCronsMsg(core.GenerateRandomID(), 2)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetCronsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetCronsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetCronsMsgEquals(t *testing.T) {
	msg := CreateGetCronsMsg(core.GenerateRandomID(), 2)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
