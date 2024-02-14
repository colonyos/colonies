package dht

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func TestValidateKey(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"/key1/key2/key3", true},                      // Valid: exactly 3 sublevels
		{"/key1/key2/key3/key4/key5", true},            // Valid: exactly 5 sublevels
		{"/key1/", false},                              // Invalid: ends with a slash
		{"/key1/key2/key3/key4/key5/key6", false},      // Invalid: more than 5 sublevels
		{"key1/key2", false},                           // Invalid: no leading slash
		{"/key1/key2//key3", false},                    // Invalid: double slash
		{"/key1/key2/key3/", false},                    // Invalid: trailing slash
		{"/key1/key2/key3/key4/key5/key6/key7", false}, // Invalid: more than 5 sublevels
		{"/", false},                                   // Invalid: no sublevels
		{"/key1key2/key3", true},                       // Valid: alphanumeric keys
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := validateKey(tt.key)
			if err != nil {
				t.Fatalf("isValidKey(%q) returned an unexpected error: %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("isValidKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestValidateRootKey(t *testing.T) {
	validKey := "/de0c93f3030f5f1e4925fd7ce597d71ed059b292663871057644e3f71b1bfba4/testkey1/testkey2"
	invalidKey := "/12D3KooWNoSCsmEyiJhVYbdFTkTVdFfH8ZhSWGdDgNDdhSGULAYb/testkey3"
	invalidKey2 := "de0c93f3030f5f1e4925fd7ce597d71ed059b292663871057644e3f71b1bfba4/testkey1/testkey2"
	id := "de0c93f3030f5f1e4925fd7ce597d71ed059b292663871057644e3f71b1bfba4"

	valid, err := validateRootKey(validKey, id)
	assert.Nil(t, err)
	assert.True(t, valid)

	valid, err = validateRootKey(invalidKey, id)
	assert.NotNil(t, err)
	assert.False(t, valid)

	valid, err = validateRootKey(invalidKey2, id)
	assert.NotNil(t, err)
	assert.False(t, valid)
}

func TestValidateValue(t *testing.T) {
	crypto := crypto.CreateCrypto()

	prvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)

	id, err := crypto.GenerateID(prvKey)
	assert.Nil(t, err)

	key := "/" + id + "/testkey1/testkey2"
	value := "testvalue"
	hash := crypto.GenerateHash(value)
	sig, err := crypto.GenerateSignature(hash, prvKey)

	kv := KV{ID: id, Key: key, Value: value, Sig: sig}
	valid, err := ValidateKV(&kv)
	assert.Nil(t, err)
	assert.True(t, valid)

	kv = KV{ID: core.GenerateRandomID(), Key: key, Value: value, Sig: sig}
	valid, err = ValidateKV(&kv)
	assert.NotNil(t, err)
	assert.False(t, valid)
}
