package validator

import (
	"errors"
)

type OwnershipMock struct {
	colonies          map[string]string
	executors         map[string]string
	users             map[string]string
	approvedExecutors map[string]bool
}

func createOwnershipMock() *OwnershipMock {
	ownership := &OwnershipMock{}
	ownership.colonies = make(map[string]string)
	ownership.executors = make(map[string]string)
	ownership.approvedExecutors = make(map[string]bool)

	return ownership
}

func (ownership *OwnershipMock) addColony(colonyID string, colonyName string) {
	ownership.colonies[colonyName] = colonyID
}

func (ownership *OwnershipMock) addExecutor(executorID string, colonyID string) {
	ownership.executors[executorID] = colonyID
}

func (ownership *OwnershipMock) addUser(userID string, colonyID string) {
	ownership.executors[userID] = colonyID
}

func (ownership *OwnershipMock) approveExecutor(executorID string, colonyID string) {
	ownership.approvedExecutors[executorID] = true
}

func (ownership *OwnershipMock) resolveColony(colonyName string) (string, error) {
	return ownership.colonies[colonyName], nil
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

func (ownership *OwnershipMock) checkIfUserBelongsToColony(userID string, colonyName string) error {
	colonyIDFromDB := ownership.users[userID]
	if colonyIDFromDB == "" {
		return errors.New("Colony does not exists")
	}
	if colonyIDFromDB != colonyName {
		return errors.New("Colony does have such a executor")
	}

	return nil
}

func (ownership *OwnershipMock) checkIfExecutorIsValid(executorID string, colonyName string, approved bool) error {
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

func (ownership *OwnershipMock) checkIfUserIsValid(userID string, colonyName string) error {
	if ownership.users[userID] == "" {
		return errors.New("User does not exists")
	}

	return nil
}
