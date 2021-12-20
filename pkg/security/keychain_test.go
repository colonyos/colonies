package security

import (
	"colonies/pkg/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeychain(t *testing.T) {
	keychain, err := CreateKeychain(".colonies_test")
	assert.Nil(t, err)

	prvKey, err := GeneratePrivateKey()
	assert.Nil(t, err)

	id := core.GenerateRandomID()
	keychain.AddPrvKey(id, prvKey)

	prvKeyFromKeychain, err := keychain.GetPrvKey(id)
	assert.Nil(t, err)
	assert.Equal(t, prvKey, prvKeyFromKeychain)

	keychain.Remove()
}
