package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddUserMsg(t *testing.T) {
	user := utils.CreateTestUser(core.GenerateRandomID(), "test_user")
	msg := CreateAddUserMsg(user)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddUserMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddUserMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddUserMsgIndent(t *testing.T) {
	user := utils.CreateTestUser(core.GenerateRandomID(), "test_user")
	msg := CreateAddUserMsg(user)

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddUserMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddUserMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddUserMsgEquals(t *testing.T) {
	user := utils.CreateTestUser(core.GenerateRandomID(), "test_user")
	msg := CreateAddUserMsg(user)

	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
