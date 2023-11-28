package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCGetSnapshotsMsg(t *testing.T) {
	msg := CreateGetSnapshotsMsg("test_colony")
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateGetSnapshotsMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateGetSnapshotsMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))
}

func TestRPCGetSnapShotsMsgEquals(t *testing.T) {
	msg := CreateGetSnapshotsMsg("test_colony")
	assert.True(t, msg.Equals(msg))
	assert.False(t, msg.Equals(nil))
}
