package server

import (
	"errors"

	"github.com/colonyos/colonies/pkg/database"
)

func resolveInitiator(
	colonyName string,
	recoveredID string,
	db database.Database) (string, error) {

	executor, err := db.GetExecutorByID(recoveredID)
	if err != nil {
		return "", err
	}

	if executor != nil {
		return executor.Name, nil
	} else {
		user, err := db.GetUserByID(colonyName, recoveredID)
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
