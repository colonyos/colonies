package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetSnapshotMsg(t *testing.T) {
	msg := CreateGetSnapshotMsg("test_colony", "test_snapshotid", "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetSnapshotMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetSnapshotMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetSnapShotMsgEquals(t *testing.T) {
	msg := CreateGetSnapshotMsg("test_colony", "test_snapshotid", "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
