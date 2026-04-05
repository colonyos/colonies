package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricsMsg(t *testing.T) {
	msg := CreateGetMetricsMsg("test_colony", "test_executor")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetMetricsMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestGetMetricsMsgIndent(t *testing.T) {
	msg := CreateGetMetricsMsg("test_colony", "test_executor")

	_, err := msg.ToJSONIndent()
	assert.Nil(t, err)
}

func TestGetMetricsMsgEquals(t *testing.T) {
	msg := CreateGetMetricsMsg("test_colony", "test_executor")

	assert.False(t, msg.Equals(nil))
}
