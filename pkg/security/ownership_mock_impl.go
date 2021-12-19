package security

import (
	"errors"
)

type OwnershipMock struct {
	colonies  map[string]bool
	computers map[string]string
}

func CreateOwnershipMock() *OwnershipMock {
	ownership := &OwnershipMock{}
	ownership.colonies = make(map[string]bool)
	ownership.computers = make(map[string]string)
	return ownership
}

func (ownership *OwnershipMock) addColony(colonyID string) {
	ownership.colonies[colonyID] = true
}

func (ownership *OwnershipMock) addComputer(computerID string, colonyID string) {
	ownership.computers[computerID] = colonyID
}

func (ownership *OwnershipMock) CheckIfColonyExists(colonyID string) error {
	colonyIDFromDB := ownership.colonies[colonyID]
	if !colonyIDFromDB {
		return errors.New("Colony does not exists")
	}

	return nil
}

func (ownership *OwnershipMock) CheckIfComputerBelongsToColony(computerID string, colonyID string) error {
	colonyIDFromDB := ownership.computers[computerID]
	if colonyIDFromDB == "" {
		return errors.New("Colony does not exists")
	}
	if colonyIDFromDB != colonyID {
		return errors.New("Colony does have such a computer")
	}

	return nil
}
