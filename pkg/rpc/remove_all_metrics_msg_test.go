package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveAllMetricsMsg(t *testing.T) {
	msg := CreateRemoveAllMetricsMsg("test_colony", "test_executor")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveAllMetricsMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestRemoveAllMetricsMsgIndent(t *testing.T) {
	msg := CreateRemoveAllMetricsMsg("test_colony", "test_executor")

	_, err := msg.ToJSONIndent()
	assert.Nil(t, err)
}

func TestRemoveAllMetricsMsgEquals(t *testing.T) {
	msg := CreateRemoveAllMetricsMsg("test_colony", "test_executor")

	assert.False(t, msg.Equals(nil))
}
