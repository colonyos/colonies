package security

import (
	"colonies/pkg/core"
	"colonies/pkg/security/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeychain(t *testing.T) {
	keychain, err := CreateKeychain(".colonies_test")
	assert.Nil(t, err)

	crypto := crypto.CreateCrypto()
	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	id := core.GenerateRandomID()
	keychain.AddPrvKey(id, prvKey)

	prvKeyFromKeychain, err := keychain.GetPrvKey(id)
	assert.Nil(t, err)
	assert.Equal(t, prvKey, prvKeyFromKeychain)

	keychain.Remove()
}
