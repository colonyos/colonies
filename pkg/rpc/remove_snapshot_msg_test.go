package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveSnapshotMsg(t *testing.T) {
	msg := CreateRemoveSnapshotMsg("test_colony", "test_snapshot", "test_name")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveSnapshotMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveSnapshotMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveSnapShotMsgEquals(t *testing.T) {
	msg := CreateRemoveSnapshotMsg("test_colony", "test_snapshot", "test_name")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
