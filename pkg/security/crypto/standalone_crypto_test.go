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
	assert.Nil(t, err)

	_, err = crypto.GenerateID(prvKey + "error")
	assert.NotNil(t, err)

	msg := "test_msg"
	signature, err := crypto.GenerateSignature(msg, prvKey)
	assert.Nil(t, err)

	msg2 := "test_msg_2"
	_, err = crypto.GenerateSignature(msg2, prvKey+"error")
	assert.NotNil(t, err)

	invalidSignature, err := crypto.GenerateSignature(msg2, prvKey)
	assert.Nil(t, err)

	recoveredID, err := crypto.RecoverID(msg, signature)
	assert.Nil(t, err)
	assert.Equal(t, recoveredID, id)

	recoveredID, err = crypto.RecoverID(msg, invalidSignature)
	assert.Nil(t, err)
	assert.NotEqual(t, recoveredID, id)
}
