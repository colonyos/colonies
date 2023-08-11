package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCDeleteSnapshotMsg(t *testing.T) {
	msg := CreateDeleteSnapshotMsg("test_colonyid", "test_snapshotid")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateDeleteSnapshotMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateDeleteSnapshotMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCDeleteSnapShotMsgEquals(t *testing.T) {
	msg := CreateDeleteSnapshotMsg("test_colonyid", "test_snapshotid")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
