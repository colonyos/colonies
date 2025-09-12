package kvstore

import (
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// ColonyDatabase Interface Implementation
// =====================================

// AddColony adds a colony to the database
func (db *KVStoreDatabase) AddColony(colony *core.Colony) error {
	if colony == nil {
		return errors.New("colony cannot be nil")
	}

	// Store colony directly
	colonyPath := fmt.Sprintf("/colonies/%s", colony.Name)
	
	// Check if colony already exists
	if db.store.Exists(colonyPath) {
		return fmt.Errorf("colony with name %s already exists", colony.Name)
	}

	err := db.store.Put(colonyPath, colony)
	if err != nil {
		return fmt.Errorf("failed to add colony %s: %w", colony.Name, err)
	}

	return nil
}

// GetColonies retrieves all colonies
func (db *KVStoreDatabase) GetColonies() ([]*core.Colony, error) {
	// Find all colonies in the /colonies directory using FindAllRecursive
	colonies, err := db.store.FindAllRecursive("/colonies", "name")
	if err != nil {
		return nil, fmt.Errorf("failed to find colonies: %w", err)
	}

	var result []*core.Colony
	for _, searchResult := range colonies {
		if colony, ok := searchResult.Value.(*core.Colony); ok {
			result = append(result, colony)
		}
	}

	return result, nil
}

// GetColonyByID retrieves a colony by ID
func (db *KVStoreDatabase) GetColonyByID(id string) (*core.Colony, error) {
	// Get all colonies first, then search by ID
	allColonies, err := db.GetColonies()
	if err != nil {
		return nil, fmt.Errorf("failed to get colonies: %w", err)
	}

	// Search through colonies for matching ID
	for _, colony := range allColonies {
		if colony.ID == id {
			return colony, nil
		}
	}

	return nil, fmt.Errorf("colony with ID %s not found", id)
}

// GetColonyByName retrieves a colony by name
func (db *KVStoreDatabase) GetColonyByName(name string) (*core.Colony, error) {
	colonyPath := fmt.Sprintf("/colonies/%s", name)
	
	if !db.store.Exists(colonyPath) {
		return nil, nil  // Return (nil, nil) when colony not found, like PostgreSQL
	}

	colonyInterface, err := db.store.Get(colonyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get colony %s: %w", name, err)
	}

	colony, ok := colonyInterface.(*core.Colony)
	if !ok {
		return nil, fmt.Errorf("stored object is not a colony")
	}

	return colony, nil
}

// RenameColony renames a colony
func (db *KVStoreDatabase) RenameColony(colonyName string, newColonyName string) error {
	oldPath := fmt.Sprintf("/colonies/%s", colonyName)
	newPath := fmt.Sprintf("/colonies/%s", newColonyName)

	// Check if old colony exists
	if !db.store.Exists(oldPath) {
		return fmt.Errorf("colony with name %s not found", colonyName)
	}

	// Check if new name already exists
	if db.store.Exists(newPath) {
		return fmt.Errorf("colony with name %s already exists", newColonyName)
	}

	// Get the colony
	colonyInterface, err := db.store.Get(oldPath)
	if err != nil {
		return fmt.Errorf("failed to get colony %s: %w", colonyName, err)
	}

	colony, ok := colonyInterface.(*core.Colony)
	if !ok {
		return fmt.Errorf("stored object is not a colony")
	}

	// Update the colony name
	colony.Name = newColonyName

	// Store at new path
	err = db.store.Put(newPath, colony)
	if err != nil {
		return fmt.Errorf("failed to store renamed colony: %w", err)
	}

	// Remove old path
	err = db.store.Delete(oldPath)
	if err != nil {
		// Try to rollback
		db.store.Delete(newPath)
		return fmt.Errorf("failed to remove old colony path: %w", err)
	}

	return nil
}

// RemoveColonyByName removes a colony by name
func (db *KVStoreDatabase) RemoveColonyByName(colonyName string) error {
	colonyPath := fmt.Sprintf("/colonies/%s", colonyName)
	
	if !db.store.Exists(colonyPath) {
		return fmt.Errorf("colony with name %s not found", colonyName)
	}

	err := db.store.Delete(colonyPath)
	if err != nil {
		return fmt.Errorf("failed to remove colony %s: %w", colonyName, err)
	}

	return nil
}

// CountColonies returns the number of colonies
func (db *KVStoreDatabase) CountColonies() (int, error) {
	colonies, err := db.GetColonies()
	if err != nil {
		return 0, err
	}

	return len(colonies), nil
}