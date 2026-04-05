package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestSetMetricMsg(t *testing.T) {
	metric := core.CreateMetric("test_colony", "test_executor", "gpu_temp", core.GAUGE, 75.5)
	msg := CreateSetMetricMsg(metric)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateSetMetricMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestSetMetricMsgIndent(t *testing.T) {
	metric := core.CreateMetric("test_colony", "test_executor", "gpu_temp", core.GAUGE, 75.5)
	msg := CreateSetMetricMsg(metric)

	_, err := msg.ToJSONIndent()
	assert.Nil(t, err)
}

func TestSetMetricMsgEquals(t *testing.T) {
	metric := core.CreateMetric("test_colony", "test_executor", "gpu_temp", core.GAUGE, 75.5)
	msg := CreateSetMetricMsg(metric)

	assert.False(t, msg.Equals(nil))
}
