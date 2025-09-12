package kvstore

import (
	"fmt"
)

// =====================================
// SecurityDatabase Interface Implementation
// =====================================

// SetServerID sets or updates the server ID
func (db *KVStoreDatabase) SetServerID(oldServerID, newServerID string) error {
	// Store server ID at a simple path (no array structure needed)
	serverIDPath := "/serverid"
	
	// If oldServerID is specified, verify it matches current
	if oldServerID != "" {
		if db.store.Exists(serverIDPath) {
			currentIDInterface, err := db.store.Get(serverIDPath)
			if err != nil {
				return fmt.Errorf("failed to get current server ID: %w", err)
			}
			if currentID, ok := currentIDInterface.(string); ok && currentID != oldServerID {
				return fmt.Errorf("old server ID does not match current ID")
			}
		}
	}

	err := db.store.Put(serverIDPath, newServerID)
	if err != nil {
		return fmt.Errorf("failed to set server ID: %w", err)
	}

	return nil
}

// GetServerID retrieves the current server ID
func (db *KVStoreDatabase) GetServerID() (string, error) {
	serverIDPath := "/serverid"
	
	if !db.store.Exists(serverIDPath) {
		return "", fmt.Errorf("server ID not found")
	}

	serverIDInterface, err := db.store.Get(serverIDPath)
	if err != nil {
		return "", fmt.Errorf("failed to get server ID: %w", err)
	}

	serverID, ok := serverIDInterface.(string)
	if !ok {
		return "", fmt.Errorf("stored server ID is not a string")
	}

	return serverID, nil
}

// ChangeColonyID changes a colony ID from old to new
func (db *KVStoreDatabase) ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error {
	// Find the colony
	colony, err := db.GetColonyByName(colonyName)
	if err != nil {
		return err
	}

	// Verify old ID matches
	if colony.ID != oldColonyID {
		return fmt.Errorf("old colony ID does not match current ID")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedColony := *colony
	updatedColony.ID = newColonyID

	// Store back
	colonyPath := fmt.Sprintf("/colonies/%s", colonyName)
	err = db.store.Put(colonyPath, &updatedColony)
	if err != nil {
		return fmt.Errorf("failed to update colony ID: %w", err)
	}

	return nil
}

// ChangeUserID changes a user ID from old to new
func (db *KVStoreDatabase) ChangeUserID(colonyName string, oldUserID, newUserID string) error {
	// Find the user by old ID
	user, err := db.GetUserByID(colonyName, oldUserID)
	if err != nil {
		return err
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedUser := *user
	updatedUser.ID = newUserID

	// Store at new path
	newUserPath := fmt.Sprintf("/users/%s", newUserID)
	err = db.store.Put(newUserPath, &updatedUser)
	if err != nil {
		return fmt.Errorf("failed to store user with new ID: %w", err)
	}

	// Remove old path
	oldUserPath := fmt.Sprintf("/users/%s", oldUserID)
	err = db.store.Delete(oldUserPath)
	if err != nil {
		// Try to rollback
		db.store.Delete(newUserPath)
		return fmt.Errorf("failed to remove old user path: %w", err)
	}

	return nil
}

// ChangeExecutorID changes an executor ID from old to new
func (db *KVStoreDatabase) ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error {
	// Find the executor by old ID
	executor, err := db.GetExecutorByID(oldExecutorID)
	if err != nil {
		return err
	}

	// Verify executor belongs to colony
	if executor.ColonyName != colonyName {
		return fmt.Errorf("executor does not belong to colony %s", colonyName)
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedExecutor := *executor
	updatedExecutor.ID = newExecutorID

	// Store at new path
	newExecutorPath := fmt.Sprintf("/executors/%s", newExecutorID)
	err = db.store.Put(newExecutorPath, &updatedExecutor)
	if err != nil {
		return fmt.Errorf("failed to store executor with new ID: %w", err)
	}

	// Remove old path
	oldExecutorPath := fmt.Sprintf("/executors/%s", oldExecutorID)
	err = db.store.Delete(oldExecutorPath)
	if err != nil {
		// Try to rollback
		db.store.Delete(newExecutorPath)
		return fmt.Errorf("failed to remove old executor path: %w", err)
	}

	return nil
}