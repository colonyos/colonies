package rpc

import (
	"colonies/pkg/core"
	"colonies/pkg/security/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineRPCMethod(t *testing.T) {
	crypto := crypto.CreateCrypto()
	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	colonyID, err := crypto.GenerateID(prvKey)
	assert.Nil(t, err)

	in := make(map[string]string)
	in["test_key"] = "test_value"
	processSpec := core.CreateProcessSpec(colonyID, []string{}, "test_runtime_type", -1, 3, 1000, 10, 1, in)

	msg := CreateSubmitProcessSpecMsg(processSpec)
	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)
	assert.Equal(t, SubmitProcessSpecMsgType, DetermineMsgType(jsonString))
}
