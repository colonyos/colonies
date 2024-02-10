package dht

import (
	"fmt"
	"regexp"

	"github.com/colonyos/colonies/internal/crypto"
)

func isValidKey(key string) (bool, error) {
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

func isValidRootKey(key, value, sig string) (bool, error) {
	// The root key must be a valid identity of the Kademlia node ID adding a key-value pair
	rootKey, err := getRootKey(key)
	if err != nil {
		return false, fmt.Errorf("Invalid root key: %v", err)
	}

	hash := crypto.GenerateHashFromString(value)
	recoveredRootKey, err := crypto.RecoveredID(hash, []byte(sig))
	if err != nil {
		return false, fmt.Errorf("Invalid signature: %v", err)
	}

	if recoveredRootKey != rootKey {
		return false, fmt.Errorf("Invalid signature")
	}

	return true, nil
}
