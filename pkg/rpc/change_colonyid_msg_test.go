package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCChangeColonyIDMsg(t *testing.T) {
	msg := CreateChangeColonyIDMsg("test_colony_name", "test_colony_id")

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateChangeColonyIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeColonyIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeColonyIDMsgIndent(t *testing.T) {
	msg := CreateChangeColonyIDMsg("test_colony_name", "test_colony_id")

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateChangeColonyIDMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateChangeColonyIDMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCChangeColonyIDMsgEquals(t *testing.T) {
	msg := CreateChangeColonyIDMsg("test_colony_name", "test_colony_id")

	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
