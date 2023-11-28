package rpc

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRPCAddFileMsg(t *testing.T) {
	file := utils.CreateTestFileWithID("test_id", "test_colony", time.Now())
	msg := CreateAddFileMsg(file)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateAddFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddFileMsgIndent(t *testing.T) {
	file := utils.CreateTestFileWithID("test_id", "test_colony", time.Now())
	msg := CreateAddFileMsg(file)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateAddFileMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateAddFileMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCAddFileMsgEquals(t *testing.T) {
	file := utils.CreateTestFileWithID("test_id", "test_colony", time.Now())
	msg := CreateAddFileMsg(file)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
