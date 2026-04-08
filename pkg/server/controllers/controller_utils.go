package controllers

import (
	"github.com/colonyos/colonies/pkg/database"
	log "github.com/sirupsen/logrus"
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
	}

	user, err := userDB.GetUserByID(colonyName, recoveredID)
	if err != nil {
		return "", err
	}
	if user != nil {
		return user.Name, nil
	}

	// Executor or user no longer exists (e.g., executor re-registered with
	// a new ID). Return the raw ID as the initiator name rather than failing.
	// This allows crons and other deferred operations to continue working
	// after executor restarts.
	log.WithFields(log.Fields{"RecoveredID": recoveredID, "ColonyName": colonyName}).Debug("Could not resolve initiator name, using ID as fallback")
	return recoveredID, nil
}
