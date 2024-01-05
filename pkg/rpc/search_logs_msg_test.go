package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCSearchLogsMsg(t *testing.T) {
	msg := CreateSearchLogsMsg("test_colony", "test_text", 1, 1)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSearchLogsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSearchLogsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSearchLogsMsgIndent(t *testing.T) {
	msg := CreateSearchLogsMsg("test_colony", "test_text", 1, 1)
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateSearchLogsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateSearchLogsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCSearchLogsMsgEquals(t *testing.T) {
	msg := CreateSearchLogsMsg("test_colony", "test_text", 1, 1)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
