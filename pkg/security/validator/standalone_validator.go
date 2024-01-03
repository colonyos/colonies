package validator

import (
	"errors"

	"github.com/colonyos/colonies/pkg/database"
)

type StandaloneValidator struct {
	ownership ownership
	db        database.Database
}

func createTestValidator(ownership ownership) *StandaloneValidator {
	return &StandaloneValidator{ownership: ownership}
}

func CreateValidator(db database.Database) *StandaloneValidator {
	return &StandaloneValidator{ownership: createOwnership(db), db: db}
}

func (validator *StandaloneValidator) RequireServerOwner(recoveredID string, serverID string) error {
	if recoveredID != serverID {
		return errors.New("Access denied, not server owner")
	}

	return nil
}

func (validator *StandaloneValidator) RequireColonyOwner(recoveredID string, colonyName string) error {
	colonyID, err := validator.ownership.resolveColony(colonyName)
	if err != nil {
		return err
	}

	if recoveredID != colonyID {
		return errors.New("Access denied, not colony owner")
	}

	return nil
}

func (validator *StandaloneValidator) RequireMembership(recoveredID string, colonyName string, approved bool) error {
	err := validator.ownership.checkIfExecutorIsValid(recoveredID, colonyName, approved)
	if err != nil {
		return validator.ownership.checkIfUserIsValid(recoveredID, colonyName)
	}

	return nil
}
