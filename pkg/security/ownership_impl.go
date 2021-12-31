package security

import (
	"colonies/pkg/core"
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

func (ownership *OwnershipImpl) CheckIfRuntimeIsValid(runtimeID string, colonyID string) error {
	colony, err := ownership.db.GetColonyByID(colonyID)
	if err != nil {
		return err
	}

	if colony == nil {
		return errors.New("Colony <" + colonyID + "> does not exists")
	}

	runtime, err := ownership.db.GetRuntimeByID(runtimeID)
	if err != nil {
		return err
	}

	if runtime == nil {
		return errors.New("Runtime with Id <" + runtimeID + "> is not a member of Colony with Id <" + colonyID + "> (Runtime does not exist)")
	}

	if runtime.ColonyID != colonyID {
		return errors.New("Runtime with Id <" + runtimeID + "> is not a member of Colony with Id <" + colonyID + ">, (Recovered Id and Colony Id missmatches)")
	}

	if runtime.Status != core.APPROVED {
		return errors.New("Runtime with Id <" + runtimeID + "> is not approved")
	}

	return nil
}
