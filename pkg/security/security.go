package security

import (
	"colonies/pkg/crypto"
	"encoding/hex"
	"errors"
)

func VerifyAPIKey(apiKey string, expectedAPIKey string) error {
	if apiKey == "" {
		return errors.New("Api-Key is missing")
	}

	if apiKey != expectedAPIKey {
		return errors.New("Invalid Api-Key")
	}

	return nil
}

func GenerateSignature(jsonString string, prvKey string) (string, error) {
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(jsonString)
	sig, err := crypto.Sign(hash, idendity.PrivateKey())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sig), nil
}

func VerifyColonyOwnership(colonyID string, data string, signature string, ownership Ownership) error {
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	hash := crypto.GenerateHash([]byte(data))
	derivedColonyID, err := crypto.RecoveredID(hash, []byte(signatureBytes))
	if err != nil {
		return err
	}

	if derivedColonyID != colonyID {
		return errors.New("invalid signature")
	}

	err = ownership.CheckIfColonyExists(colonyID)
	if err != nil {
		return err
	}

	return nil
}
