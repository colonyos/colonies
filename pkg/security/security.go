package security

import (
	"colonies/pkg/crypto"
	"encoding/hex"
	"errors"
)

func GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", nil
	}

	return identify.PrivateKeyAsHex(), nil
}

func GenerateID(privateKey string) (string, error) {
	identify, err := crypto.CreateIdendityFromString(privateKey)
	if err != nil {
		return "", nil
	}

	return identify.ID(), nil
}

func GenerateSignature(jsonString string, prvKey string) (string, error) { // TODO: unittest
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(jsonString)
	signatureBytes, err := crypto.Sign(hash, idendity.PrivateKey())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signatureBytes), nil
}

func RecoverID(jsonString string, signature string) (string, error) {
	signatureString, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(jsonString)
	return crypto.RecoveredID(hash, []byte(signatureString))
}

func RequireRoot(rootPassword string, expectedRootPassword string) error {
	if rootPassword == "" {
		return errors.New("Root password is missing")
	}

	if rootPassword != expectedRootPassword {
		return errors.New("Invalid root password")
	}

	return nil
}

func RequireColonyOwner(recoveredID string, colonyID string, ownership Ownership) error {
	if recoveredID != colonyID {
		return errors.New("RecoveredID does not match Colony with Id <" + colonyID + ">")
	}

	return ownership.CheckIfColonyExists(colonyID)
}

func RequireRuntimeMembership(runtimeID string, colonyID string, ownership Ownership) error {
	return ownership.CheckIfRuntimeIsValid(runtimeID, colonyID)
}
