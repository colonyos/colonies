package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoveredID(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	recoveredID, err := RecoveredID(hash, sig)
	assert.Nil(t, err)
	assert.Equal(t, idendity.ID(), recoveredID)
}

func TestRecoverFromStrings(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	hash2, err := CreateHashFromString(hash.String())
	assert.Nil(t, err)

	recoveredID, err := RecoveredID(hash2, sig)
	assert.Nil(t, err)
	assert.Equal(t, idendity.ID(), recoveredID)
}

func TestRecoverFromStringsInvalidHash(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	hash = GenerateHash([]byte("blablabla"))

	recoveredID, err := RecoveredID(hash, sig)
	assert.Nil(t, err)
	assert.NotEqual(t, idendity.ID(), recoveredID)
}

func TestRecoverPublicKey(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	hash := GenerateHash([]byte("test"))

	sig, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	pub, err := RecoverPublicKey(hash, sig)
	assert.Nil(t, err)
	assert.Equal(t, idendity.PublicKeyAsHex(), hex.EncodeToString(pub))
}

func TestRecoverPublicKeyInvalidSignature(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	idendity2, err := CreateIdendity()
	assert.Nil(t, err)

	hash := GenerateHash([]byte("test"))

	sig, err := Sign(hash, idendity2.PrivateKey())
	assert.Nil(t, err)

	pub, err := RecoverPublicKey(hash, sig)
	assert.Nil(t, err)
	assert.NotEqual(t, idendity.PublicKeyAsHex(), hex.EncodeToString(pub))
}

func TestSignAndVerify(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	decPub, err := hex.DecodeString(idendity.PublicKeyAsHex())
	assert.Nil(t, err)

	ok, err := Verify(decPub, hash, sig)
	assert.Nil(t, err)
	assert.True(t, ok)
}

func TestSignAndVerifyInvalidPubKey(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	idendity2, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	sig, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	decPub, err := hex.DecodeString(idendity2.PublicKeyAsHex())
	assert.Nil(t, err)

	ok, err := Verify(decPub, hash, sig)
	assert.Nil(t, err)
	assert.False(t, ok)
}
