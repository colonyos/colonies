package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCCreateSnapshotMsg(t *testing.T) {
	msg := CreateCreateSnapshotMsg("test_colony", "test_label", "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateCreateSnapshotMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateCreateSnapshotMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCCreateSnapShotMsgEquals(t *testing.T) {
	msg := CreateCreateSnapshotMsg("test_colony", "test_label", "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
