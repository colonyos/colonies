package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterMsg(t *testing.T) {
	// test cluster message
	msg := &ClusterMsg{
		MsgType:    PingRequestMsgType,
		ID:         "2",
		Originator: "replica1",
		Recipient:  "replica2",
		Data:       []byte{1, 2, 3},
	}

	serializedMsg, err := msg.Serialize()
	assert.Nil(t, err)

	deserializedMsg, err := DeserializeClusterMsg(serializedMsg)
	assert.Nil(t, err)

	assert.Equal(t, msg.Equals(deserializedMsg), true)
}
