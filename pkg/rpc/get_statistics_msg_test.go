package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetStatisticsMsg(t *testing.T) {
	msg := CreateGetStatisticsMsg()
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetStatisticsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetStatisticsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetStatisticsMsgIndent(t *testing.T) {
	msg := CreateGetStatisticsMsg()
	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateGetStatisticsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetStatisticsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetStatisticsMsgEquals(t *testing.T) {
	msg := CreateGetStatisticsMsg()
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
