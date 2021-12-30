package security

import (
	"colonies/pkg/crypto"
	"encoding/hex"
	"errors"
)

// TODO: Pending or disapproved runtimes should be blocked!

func GenerateID(privateKey string) (string, error) {
	identify, err := crypto.CreateIdendityFromString(privateKey)
	if err != nil {
		return "", nil
	}

	return identify.ID(), nil
}

func GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", nil
	}

	return identify.PrivateKeyAsHex(), nil
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

func RecoverID(jsonString string, signature string) (string, error) { // TODO: unittest
	signatureString, err := hex.DecodeString(signature)
	hash := crypto.GenerateHashFromString(jsonString)
	derivedID, err := crypto.RecoveredID(hash, []byte(signatureString))
	if err != nil {
		return "", err
	}

	return derivedID, nil
}

func VerifyRoot(rootPassword string, expectedRootPassword string) error {
	if rootPassword == "" {
		return errors.New("Root password is missing")
	}

	if rootPassword != expectedRootPassword {
		return errors.New("Invalid root password")
	}

	return nil
}

func VerifyColonyOwner(recoveredID string, colonyID string, ownership Ownership) error {
	if recoveredID != colonyID {
		return errors.New("RecoveredID does not match Colony with Id <" + colonyID + ">")
	}

	err := ownership.CheckIfColonyExists(colonyID)
	if err != nil {
		return err
	}

	return nil
}

func VerifyRuntimeMembership(runtimeID string, colonyID string, ownership Ownership) error {
	return ownership.CheckIfRuntimeBelongsToColony(runtimeID, colonyID)
}
