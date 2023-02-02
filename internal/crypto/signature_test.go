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

	signatureBytes, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	recoveredID, err := RecoveredID(hash, []byte(string(signatureBytes)+"too_large_signature"))
	assert.NotNil(t, err)

	recoveredID, err = RecoveredID(hash, signatureBytes)
	assert.Nil(t, err)
	assert.Equal(t, idendity.ID(), recoveredID)
}

func TestRecoverFromStrings(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	signatureBytes, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	hash2, err := CreateHashFromString(hash.String())
	assert.Nil(t, err)

	recoveredID, err := RecoveredID(hash2, signatureBytes)
	assert.Nil(t, err)
	assert.Equal(t, idendity.ID(), recoveredID)
}

func TestRecoverFromStringsInvalidHash(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	signatureBytes, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	hash = GenerateHash([]byte("blablabla"))

	recoveredID, err := RecoveredID(hash, signatureBytes)
	assert.Nil(t, err)
	assert.NotEqual(t, idendity.ID(), recoveredID)
}

func TestRecoverPublicKey(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	hash := GenerateHash([]byte("test"))

	signatureBytes, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	pub, err := RecoverPublicKey(hash, signatureBytes)
	assert.Nil(t, err)
	assert.Equal(t, idendity.PublicKeyAsHex(), hex.EncodeToString(pub))
}

func TestRecoverPublicKeyInvalidSignature(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	idendity2, err := CreateIdendity()
	assert.Nil(t, err)

	hash := GenerateHash([]byte("test"))

	signatureBytes, err := Sign(hash, idendity2.PrivateKey())
	assert.Nil(t, err)

	pub, err := RecoverPublicKey(hash, signatureBytes)
	assert.Nil(t, err)
	assert.NotEqual(t, idendity.PublicKeyAsHex(), hex.EncodeToString(pub))
}

func TestSignAndVerify(t *testing.T) {
	idendity, err := CreateIdendity()
	assert.Nil(t, err)

	msg := "test"
	hash := GenerateHashFromString(msg)

	signatureBytes, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	decPub, err := hex.DecodeString(idendity.PublicKeyAsHex())
	assert.Nil(t, err)

	ok, err := Verify(decPub, hash, signatureBytes)
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

	signatureBytes, err := Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)

	decPub, err := hex.DecodeString(idendity2.PublicKeyAsHex())
	assert.Nil(t, err)

	ok, err := Verify(decPub, hash, signatureBytes)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestInterop(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	idendity, err := CreateIdendityFromString(prvKey)
	assert.Nil(t, err)

	hash := GenerateHashFromString("hello")

	// signature, err := Sign(hash, idendity.PrivateKey())
	// assert.Nil(t, err)
	// signatureStr := hex.EncodeToString(signature)
	// fmt.Println("prvkey: " + idendity.PrivateKeyAsHex())
	// fmt.Println("pubkey: " + idendity.PublicKeyAsHex())
	// fmt.Println("id: " + idendity.ID())
	// fmt.Println("digest: " + hash.String())
	// fmt.Println("signature: " + string(signatureStr))

	signatureHex := "997eca36736d465e0e8d64e6d657ff4c939c8f5cad4272797ea0fe372bfd8d0953d21b3d06ded5dd80aee8cfa3a9be7ce615ce690eb64184fe15962943fe541300"
	signatureBytes, err := hex.DecodeString(signatureHex)
	recoveredID, err := RecoveredID(hash, signatureBytes)
	assert.Nil(t, err)
	assert.Equal(t, recoveredID, idendity.ID())
}
