package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricMsg(t *testing.T) {
	msg := CreateGetMetricMsg("test_colony", "test_executor", "gpu_temp")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetMetricMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestGetMetricMsgIndent(t *testing.T) {
	msg := CreateGetMetricMsg("test_colony", "test_executor", "gpu_temp")

	_, err := msg.ToJSONIndent()
	assert.Nil(t, err)
}

func TestGetMetricMsgEquals(t *testing.T) {
	msg := CreateGetMetricMsg("test_colony", "test_executor", "gpu_temp")

	assert.False(t, msg.Equals(nil))
}
