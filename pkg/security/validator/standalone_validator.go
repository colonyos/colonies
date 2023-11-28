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
		return errors.New("Access denied, not Server owner")
	}

	return nil
}

func (validator *StandaloneValidator) RequireColonyOwner(recoveredID string, colonyID string) error {
	if recoveredID != colonyID {
		return errors.New("Access denied, not Colony owner")
	}

	return validator.ownership.checkIfColonyExists(colonyID)
}

func (validator *StandaloneValidator) RequireMembership(recoveredID string, colonyID string, approved bool) error {
	err := validator.ownership.checkIfExecutorIsValid(recoveredID, colonyID, approved)
	if err != nil {
		return validator.ownership.checkIfUserIsValid(recoveredID, colonyID)
	}

	return nil
}
