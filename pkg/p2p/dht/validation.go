package dht

import (
	"fmt"
	"regexp"

	"github.com/colonyos/colonies/pkg/security/crypto"
)

func ValidateKV(kv *KV) (bool, error) {
	if kv == nil {
		return false, fmt.Errorf("Invalid key-value pair, kv is nil")
	}

	valid, err := validateValue(kv)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, err
	}

	valid, err = validateRootKey(kv.Key, kv.ID)
	if err != nil {
		return false, err
	}
	if !valid {
		return false, err
	}

	return true, nil
}

func validateRootKey(key string, id string) (bool, error) {
	rootKey, err := getRootKey(key) // Will call validateKey
	if err != nil {
		return false, err
	}

	if len(rootKey) != 64 {
		return false, fmt.Errorf("Invalid root key length")
	}

	if id != rootKey {
		return false, fmt.Errorf("Invalid root key")
	}

	return true, nil
}

func validateKey(key string) (bool, error) {
	// Define the regular expression pattern for validation
	// ^/                  : The string must start with a single slash
	// ([a-zA-Z0-9]+/)     : Must have one or more alphanumeric characters followed by a slash
	// {0,4}               : This group can repeat from 0 to 4 times to allow up to 5 sublevels
	// [a-zA-Z0-9]+$       : Must end with one or more alphanumeric characters (no trailing slash)
	pattern := `^/([a-zA-Z0-9]+/){0,4}[a-zA-Z0-9]+$`

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("Invalid regex pattern: %v", err)
	}

	isValid := regex.MatchString(key)

	return isValid, nil
}

func validateValue(kv *KV) (bool, error) {
	crypto := crypto.CreateCrypto()

	hash := crypto.GenerateHash(kv.Value)

	recoveredID, err := crypto.RecoverID(hash, kv.Sig)
	if err != nil {
		return false, fmt.Errorf("Invalid signature: %v", err)
	}

	if recoveredID != kv.ID {
		return false, fmt.Errorf("Invalid signature")
	}

	return true, nil
}
