package security

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
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
