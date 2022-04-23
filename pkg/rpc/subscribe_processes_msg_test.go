package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCSubscribeProcessesMsg(t *testing.T) {
	msg := CreateSubscribeProcessesMsg("test_runtime_type", 1, 2)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeProcessesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubscribeProcessesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubscribeProcessesMsgIndent(t *testing.T) {
	msg := CreateSubscribeProcessesMsg("test_runtime_type", 1, 2)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSubscribeProcessesMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSubscribeProcessesMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSubscribeProcessesMsgEquals(t *testing.T) {
	msg := CreateSubscribeProcessesMsg("test_runtime_type", 1, 2)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
