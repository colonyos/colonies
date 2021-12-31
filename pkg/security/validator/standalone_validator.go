package validator

import (
	"colonies/internal/crypto"
	"colonies/pkg/database"
	"errors"
)

type StandaloneValidator struct {
	ownership ownership
}

func createValidatorTest(ownership ownership) *StandaloneValidator {
	return &StandaloneValidator{ownership: ownership}
}

func CreateValidator(db database.Database) *StandaloneValidator {
	return &StandaloneValidator{ownership: createOwnership(db)}
}

func (validator *StandaloneValidator) GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", nil
	}

	return identify.PrivateKeyAsHex(), nil
}

func (validator *StandaloneValidator) RequireRoot(rootPassword string, expectedRootPassword string) error {
	if rootPassword == "" {
		return errors.New("Root password is missing")
	}

	if rootPassword != expectedRootPassword {
		return errors.New("Invalid root password")
	}

	return nil
}

func (validator *StandaloneValidator) RequireColonyOwner(recoveredID string, colonyID string) error {
	if recoveredID != colonyID {
		return errors.New("RecoveredID does not match Colony with Id <" + colonyID + ">")
	}

	return validator.ownership.checkIfColonyExists(colonyID)
}

func (validator *StandaloneValidator) RequireRuntimeMembership(runtimeID string, colonyID string) error {
	return validator.ownership.checkIfRuntimeIsValid(runtimeID, colonyID)
}
