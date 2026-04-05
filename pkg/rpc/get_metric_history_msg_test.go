package rpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMetricHistoryMsg(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	msg := CreateGetMetricHistoryMsg("test_colony", "test_executor", "tokens_used", 3, from, to)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetMetricHistoryMsgFromJSON(jsonString)
	assert.Nil(t, err)
	assert.True(t, msg.Equals(msg2))
}

func TestGetMetricHistoryMsgIndent(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	msg := CreateGetMetricHistoryMsg("test_colony", "test_executor", "tokens_used", 3, from, to)

	_, err := msg.ToJSONIndent()
	assert.Nil(t, err)
}

func TestGetMetricHistoryMsgEquals(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	msg := CreateGetMetricHistoryMsg("test_colony", "test_executor", "tokens_used", 3, from, to)

	assert.False(t, msg.Equals(nil))
}
