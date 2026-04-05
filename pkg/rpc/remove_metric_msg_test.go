package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveMetricMsg(t *testing.T) {
	msg := CreateRemoveMetricMsg("test_colony", "test_executor", "gpu_temp")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveMetricMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRemoveMetricMsgIndent(t *testing.T) {
	msg := CreateRemoveMetricMsg("test_colony", "test_executor", "gpu_temp")

	_, err := msg.ToJSONIndent()
	assert.Nil(t, err)
}

func TestRemoveMetricMsgEquals(t *testing.T) {
	msg := CreateRemoveMetricMsg("test_colony", "test_executor", "gpu_temp")

	assert.False(t, msg.Equals(nil))
}
