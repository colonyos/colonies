package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileLabelsMsg(t *testing.T) {
	msg := CreateGetFileLabelsMsg("test_colony", "test_label", true)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetFileLabelsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFileLabelsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFileLabelsMsgIndent(t *testing.T) {
	msg := CreateGetAllFileLabelsMsg("test_colony")
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetFileLabelsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetFileLabelsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetFileLAbelsMsgEquals(t *testing.T) {
	msg := CreateGetFileLabelsMsg("test_colony", "test_label", false)
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
