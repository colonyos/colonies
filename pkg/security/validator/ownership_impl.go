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
		return errors.New("Executor with Id <" + executorID + "> is not a member of Colony with Id <" + colonyID + ">, (Failed to receover Colony Id)")
	}

	if approved {
		if executor.State != core.APPROVED {
			return errors.New("Executor with Id <" + executorID + "> is not approved")
		}
	}

	return nil
}

func (ownership *ownershipImpl) checkIfUserIsValid(userID string, colonyID string) error {
	colony, err := ownership.db.GetColonyByID(colonyID)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony <" + colonyID + "> does not exists")
	}

	user, err := ownership.db.GetUserByID(colony.Name, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("User or Executor with Id <" + userID + "> is not a member of Colony with Id <" + colonyID + "> (User does not exist)")
	}

	if user.ColonyName != colony.Name {
		return errors.New("User with Executor Id <" + user.ID + "> is not a member of Colony with Name <" + colony.Name + ">, (Failed to receover Colony Id)")
	}

	return nil
}
