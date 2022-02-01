package validator

import (
	"errors"

	"github.com/colonyos/colonies/pkg/database"
)

type StandaloneValidator struct {
	ownership ownership
}

func createTestValidator(ownership ownership) *StandaloneValidator {
	return &StandaloneValidator{ownership: ownership}
}

func CreateValidator(db database.Database) *StandaloneValidator {
	return &StandaloneValidator{ownership: createOwnership(db)}
}

func (validator *StandaloneValidator) RequireServerOwner(recoveredID string, serverID string) error {
	if recoveredID != serverID {
		return errors.New("RecoveredID does not match Server Id")
	}

	return nil
}

func (validator *StandaloneValidator) RequireColonyOwner(recoveredID string, colonyID string) error {
	if recoveredID != colonyID {
		return errors.New("RecoveredID does not match Colony Id")
	}

	return validator.ownership.checkIfColonyExists(colonyID)
}

func (validator *StandaloneValidator) RequireRuntimeMembership(recoveredID string, colonyID string, approved bool) error {
	return validator.ownership.checkIfRuntimeIsValid(recoveredID, colonyID, approved)
}
