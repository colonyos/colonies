package rpc

import (
	"colonies/pkg/core"
	"colonies/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineRPCMethod(t *testing.T) {
	colonyIdendity, err := crypto.CreateIdendity()
	assert.Nil(t, err)
	colonyID := colonyIdendity.ID()

	in := make(map[string]string)
	in["test_key"] = "test_value"
	processSpec := core.CreateProcessSpec(colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, in)

	msg := CreateSubmitProcessSpecMsg(processSpec)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)
	assert.Equal(t, SubmitProcessSpecMsgType, DetermineMsgType(jsonString))
}
