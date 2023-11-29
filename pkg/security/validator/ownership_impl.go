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

func (ownership *ownershipImpl) resolveColony(colonyName string) (string, error) {
	if colonyName == "" {
		return "", errors.New("Empty Colony name")
	}

	colony, err := ownership.db.GetColonyByName(colonyName)
	if err != nil {
		return "", err
	}

	if colony == nil {
		return "", errors.New("Colony with name <" + colonyName + "> does not exists")
	}

	if colony.ID == "" {
		return "", errors.New("No Colony Id found")
	}

	return colony.ID, nil
}

func (ownership *ownershipImpl) checkIfExecutorIsValid(executorID string, colonyName string, approved bool) error {
	colony, err := ownership.db.GetColonyByName(colonyName)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony <" + colonyName + "> does not exists")
	}

	executor, err := ownership.db.GetExecutorByID(executorID)
	if err != nil {
		return err
	}

	if executor == nil {
		return errors.New("Access denied, not a member of Colony with name <" + colonyName + ">")
	}

	if executor.ColonyName != colony.Name {
		return errors.New("Access denied, not a member of Colony with name <" + colonyName + ">")
	}

	if approved {
		if executor.State != core.APPROVED {
			return errors.New("Access denied, Executor with Id <" + executorID + "> is not approved")
		}
	}

	return nil
}

func (ownership *ownershipImpl) checkIfUserIsValid(userID string, colonyName string) error {
	colony, err := ownership.db.GetColonyByName(colonyName)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony with name <" + colonyName + "> does not exists")
	}

	user, err := ownership.db.GetUserByID(colony.Name, userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("Access denied, not a member of Colony with name <" + colony.Name + ">")
	}

	if user.ColonyName != colony.Name {
		return errors.New("Access denied, not a member of Colony with name <" + colony.Name + ">")
	}

	return nil
}
