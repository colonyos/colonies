package validator

import (
	"errors"
)

type OwnershipMock struct {
	colonies         map[string]bool
	runtimes         map[string]string
	approvedRuntimes map[string]bool
}

func createOwnershipMock() *OwnershipMock {
	ownership := &OwnershipMock{}
	ownership.colonies = make(map[string]bool)
	ownership.runtimes = make(map[string]string)
	ownership.approvedRuntimes = make(map[string]bool)

	return ownership
}

func (ownership *OwnershipMock) addColony(colonyID string) {
	ownership.colonies[colonyID] = true
}

func (ownership *OwnershipMock) addRuntime(runtimeID string, colonyID string) {
	ownership.runtimes[runtimeID] = colonyID
}

func (ownership *OwnershipMock) approveRuntime(runtimeID string, colonyID string) {
	ownership.approvedRuntimes[runtimeID] = true
}

func (ownership *OwnershipMock) checkIfColonyExists(colonyID string) error {
	colonyIDFromDB := ownership.colonies[colonyID]
	if !colonyIDFromDB {
		return errors.New("Colony does not exists")
	}

	return nil
}

func (ownership *OwnershipMock) checkIfRuntimeBelongsToColony(runtimeID string, colonyID string) error {
	colonyIDFromDB := ownership.runtimes[runtimeID]
	if colonyIDFromDB == "" {
		return errors.New("Colony does not exists")
	}
	if colonyIDFromDB != colonyID {
		return errors.New("Colony does have such a runtime")
	}

	return nil
}

func (ownership *OwnershipMock) checkIfRuntimeIsValid(runtimeID string, colonyID string, approved bool) error {
	if ownership.runtimes[runtimeID] == "" {
		return errors.New("Runtime does not exists")
	}

	if approved {
		if ownership.approvedRuntimes[runtimeID] == false {
			return errors.New("Runtime is not approved")
		}
	}

	return nil
}
