package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCRemoveAllSnapshotsMsg(t *testing.T) {
	msg := CreateRemoveAllSnapshotsMsg("test_colony")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRemoveAllSnapshotsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRemoveAllSnapshotsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCRemoveAllSnapShotsMsgEquals(t *testing.T) {
	msg := CreateRemoveAllSnapshotsMsg("test_colony")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
