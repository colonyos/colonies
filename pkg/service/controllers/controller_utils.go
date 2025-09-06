package controllers

import (
	"errors"

	"github.com/colonyos/colonies/pkg/database"
)

func resolveInitiator(
	colonyName string,
	recoveredID string,
	executorDB database.ExecutorDatabase,
	userDB database.UserDatabase) (string, error) {

	executor, err := executorDB.GetExecutorByID(recoveredID)
	if err != nil {
		return "", err
	}

	if executor != nil {
		return executor.Name, nil
	} else {
		user, err := userDB.GetUserByID(colonyName, recoveredID)
		if err != nil {
			return "", err
		}
		if user != nil {
			return user.Name, nil
		} else {
			return "", errors.New("Could not derive InitiatorName")
		}
	}
}