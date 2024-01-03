package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCChangeExecutorIDMsg(t *testing.T) {
	msg := CreateChangeExecutorIDMsg("test_colony_name", "test_executor_id")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateChangeExecutorIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeExecutorIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeExecutorIDMsgIndent(t *testing.T) {
	msg := CreateChangeExecutorIDMsg("test_colony_name", "test_executor_id")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateChangeExecutorIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeExecutorIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeExecutorIDMsgEquals(t *testing.T) {
	msg := CreateChangeExecutorIDMsg("test_colony_name", "test_executor_id")

	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
