package dht

import (
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/colonyos/colonies/internal/crypto"
)

func ValidateKey(key string) (bool, error) {
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

func ValidateValue(kv KV) (bool, error) {
	hash := crypto.GenerateHashFromString(kv.Value)

	sigBytes, err := hex.DecodeString(kv.Sig)
	recoveredID, err := crypto.RecoveredID(hash, []byte(sigBytes))
	if err != nil {
		return false, fmt.Errorf("Invalid signature: %v", err)
	}

	if recoveredID != kv.ID {
		return false, fmt.Errorf("Invalid signature")
	}

	return true, nil
}
