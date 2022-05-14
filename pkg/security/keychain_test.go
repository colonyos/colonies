package security

import (
	"os"
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

func TestKeychainFailure(t *testing.T) {
	keychain, err := CreateKeychain(".colonies_test")
	assert.Nil(t, err)

	_, err = CreateKeychain(".colonies_test")
	assert.Nil(t, err)

	keychain.Remove()

	// Test that is actually deleted
	_, err = os.Stat(keychain.dirName)
	assert.NotNil(t, err)

	// Create a file with the same name so that keychain cannot create a directory
	_, err = os.Create(keychain.dirName)
	assert.Nil(t, err)

	_, err = CreateKeychain(".colonies_test")
	assert.NotNil(t, err)

	err = os.Remove(keychain.dirName)
	assert.Nil(t, err)
}
