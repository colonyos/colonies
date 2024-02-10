package dht

import (
	"errors"
	"strings"
)

func getRootKey(key string) (string, error) {
	valid, err := isValidKey(key)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", errors.New("Invalid key format. Expected format: '/key1/key2/.../keyN' with 1 to 5 alphanumeric sublevels and no trailing slash.")
	}

	parts := strings.FieldsFunc(key, func(r rune) bool {
		return r == '/'
	})

	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", errors.New("No prefix found")
}
