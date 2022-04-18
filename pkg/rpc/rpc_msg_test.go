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

	msg, err := CreateRPCMsg("test_method", "test_payload", prvKey)
	assert.Nil(t, err)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRPCMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRPCMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))

	assert.Equal(t, msg.DecodePayload(), "test_payload")
}

func TestRPCMsgIndent(t *testing.T) {
	crypto := crypto.CreateCrypto()
	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	msg, err := CreateRPCMsg("test_method", "test_payload", prvKey)
	assert.Nil(t, err)

	jsonString, err := msg.ToJSONIndent()
	assert.Nil(t, err)

	msg2, err := CreateRPCMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRPCMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))

	assert.Equal(t, msg.DecodePayload(), "test_payload")
}

func TestRPCMsgInsecure(t *testing.T) {
	msg, err := CreateInsecureRPCMsg("test_method", "test_payload")
	assert.Nil(t, err)

	jsonString, err := msg.ToJSON()
	assert.Nil(t, err)

	msg2, err := CreateRPCMsgFromJSON(jsonString + "error")
	assert.NotNil(t, err)

	msg2, err = CreateRPCMsgFromJSON(jsonString)
	assert.Nil(t, err)

	assert.True(t, msg.Equals(msg2))

	assert.Equal(t, msg.DecodePayload(), "test_payload")
}
