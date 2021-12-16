package security

import (
	"errors"
)

type OwnershipMock struct {
	colonies map[string]bool
	workers  map[string]string
}

func CreateOwnershipMock() *OwnershipMock {
	ownership := &OwnershipMock{}
	ownership.colonies = make(map[string]bool)
	ownership.workers = make(map[string]string)
	return ownership
}

func (ownership *OwnershipMock) AddColony(colonyID string) {
	ownership.colonies[colonyID] = true
}

func (ownership *OwnershipMock) AddWorker(colonyID string, workerID string) {
	ownership.workers[workerID] = colonyID
}

func (ownership *OwnershipMock) CheckIfColonyExists(colonyID string) error {
	colonyIDFromDB := ownership.colonies[colonyID]
	if !colonyIDFromDB {
		return errors.New("colony does not exists")
	}

	return nil
}

func (ownership *OwnershipMock) CheckIfColonyHasWorker(colonyID string, workerID string) error {
	colonyIDFromDB := ownership.workers[workerID]
	if colonyIDFromDB == "" {
		return errors.New("colony does not exists")
	}
	if colonyIDFromDB != colonyID {
		return errors.New("colony does have such a worker")
	}

	return nil
}
