package rpc

import (
	"testing"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestRPCMsg(t *testing.T) {
	crypto := crypto.CreateCrypto()
	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	rpcMsg, err := CreateRPCMsg("test_method", "test_payload", prvKey)
	assert.Nil(t, err)

	_, err = rpcMsg.ToJSONIndent()
	assert.Nil(t, err)

	// TODO
}
