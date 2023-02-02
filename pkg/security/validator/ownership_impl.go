package validator

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
)

type ownershipImpl struct {
	db database.Database
}

func createOwnership(db database.Database) *ownershipImpl {
	ownership := &ownershipImpl{}
	ownership.db = db
	return ownership
}

func (ownership *ownershipImpl) checkIfColonyExists(colonyID string) error {
	colony, err := ownership.db.GetColonyByID(colonyID)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony <" + colonyID + "> does not exists")
	}

	return nil
}

func (ownership *ownershipImpl) checkIfExecutorIsValid(executorID string, colonyID string, approved bool) error {
	colony, err := ownership.db.GetColonyByID(colonyID)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony <" + colonyID + "> does not exists")
	}

	executor, err := ownership.db.GetExecutorByID(executorID)
	if err != nil {
		return err
	}

	if executor == nil {
		return errors.New("Executor with Id <" + executorID + "> is not a member of Colony with Id <" + colonyID + "> (Executor does not exist)")
	}

	if executor.ColonyID != colonyID {
		return errors.New("Executor with Id <" + executorID + "> is not a member of Colony with Id <" + colonyID + ">, (Recovered Id and Colony Id missmatches)")
	}

	if approved {
		if executor.State != core.APPROVED {
			return errors.New("Executor with Id <" + executorID + "> is not approved")
		}
	}

	return nil
}
