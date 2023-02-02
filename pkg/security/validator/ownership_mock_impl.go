package validator

import (
	"errors"
)

type OwnershipMock struct {
	colonies          map[string]bool
	executors         map[string]string
	approvedExecutors map[string]bool
}

func createOwnershipMock() *OwnershipMock {
	ownership := &OwnershipMock{}
	ownership.colonies = make(map[string]bool)
	ownership.executors = make(map[string]string)
	ownership.approvedExecutors = make(map[string]bool)

	return ownership
}

func (ownership *OwnershipMock) addColony(colonyID string) {
	ownership.colonies[colonyID] = true
}

func (ownership *OwnershipMock) addExecutor(executorID string, colonyID string) {
	ownership.executors[executorID] = colonyID
}

func (ownership *OwnershipMock) approveExecutor(executorID string, colonyID string) {
	ownership.approvedExecutors[executorID] = true
}

func (ownership *OwnershipMock) checkIfColonyExists(colonyID string) error {
	colonyIDFromDB := ownership.colonies[colonyID]
	if !colonyIDFromDB {
		return errors.New("Colony does not exists")
	}

	return nil
}

func (ownership *OwnershipMock) checkIfExecutorBelongsToColony(executorID string, colonyID string) error {
	colonyIDFromDB := ownership.executors[executorID]
	if colonyIDFromDB == "" {
		return errors.New("Colony does not exists")
	}
	if colonyIDFromDB != colonyID {
		return errors.New("Colony does have such a executor")
	}

	return nil
}

func (ownership *OwnershipMock) checkIfExecutorIsValid(executorID string, colonyID string, approved bool) error {
	if ownership.executors[executorID] == "" {
		return errors.New("Executor does not exists")
	}

	if approved {
		if ownership.approvedExecutors[executorID] == false {
			return errors.New("Executor is not approved")
		}
	}

	return nil
}
