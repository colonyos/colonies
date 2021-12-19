package security

import (
	"colonies/pkg/database"
	"errors"
)

type OwnershipImpl struct {
	db database.Database
}

func CreateOwnership(db database.Database) *OwnershipImpl {
	ownership := &OwnershipImpl{}
	ownership.db = db
	return ownership
}

func (ownership *OwnershipImpl) CheckIfColonyExists(colonyID string) error {
	colony, err := ownership.db.GetColonyByID(colonyID)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony <" + colonyID + "> does not exists")
	}

	return nil
}

func (ownership *OwnershipImpl) CheckIfComputerBelongsToColony(computerID string, colonyID string) error {
	computer, err := ownership.db.GetComputerByID(computerID)
	if err != nil {
		return err
	}

	if computer == nil {
		return errors.New("Computer not found <" + computerID + ">")
	}

	if computer.ColonyID() != colonyID {
		return errors.New("Computer " + "<" + computerID + "> not member of colony <" + colonyID + ">")
	}

	return nil
}
