package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverID(t *testing.T) {
	crypto := CreateCrypto()

	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	id, err := crypto.GenerateID(prvKey)

	msg := "test_msg"
	signature, err := crypto.GenerateSignature(msg, prvKey)
	assert.Nil(t, err)

	recoveredID, err := crypto.RecoverID(msg, signature)
	assert.Nil(t, err)
	assert.Equal(t, recoveredID, id)
}
