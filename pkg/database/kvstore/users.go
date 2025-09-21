package kvstore

import (
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// UserDatabase Interface Implementation  
// =====================================

// AddUser adds a user to the database
func (db *KVStoreDatabase) AddUser(user *core.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}

	// Store user directly at path
	userPath := fmt.Sprintf("/users/%s", user.ID)

	// Check if user already exists
	if db.store.Exists(userPath) {
		return fmt.Errorf("user with ID %s already exists in colony %s", user.ID, user.ColonyName)
	}

	err := db.store.Put(userPath, user)
	if err != nil {
		return fmt.Errorf("failed to add user %s: %w", user.ID, err)
	}

	return nil
}

// GetUsersByColonyName retrieves all users for a colony
func (db *KVStoreDatabase) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	// Search for all users in the users directory
	users, err := db.store.FindRecursive("/users", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find users for colony %s: %w", colonyName, err)
	}

	var result []*core.User
	for _, searchResult := range users {
		if user, ok := searchResult.Value.(*core.User); ok {
			result = append(result, user)
		}
	}

	return result, nil
}

// GetUserByID retrieves a user by colony name and user ID
func (db *KVStoreDatabase) GetUserByID(colonyName string, userID string) (*core.User, error) {
	userPath := fmt.Sprintf("/users/%s", userID)
	
	if !db.store.Exists(userPath) {
		// Return (nil, nil) when user not found, like PostgreSQL
		return nil, nil
	}

	userInterface, err := db.store.Get(userPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
	}

	user, ok := userInterface.(*core.User)
	if !ok {
		return nil, fmt.Errorf("stored object is not a user")
	}

	// Check if user belongs to the specified colony
	if user.ColonyName != colonyName {
		// Return (nil, nil) when user not in specified colony, like PostgreSQL
		return nil, nil
	}

	return user, nil
}

// GetUserByName retrieves a user by colony name and user name
func (db *KVStoreDatabase) GetUserByName(colonyName string, name string) (*core.User, error) {
	// Search for user by name within the users directory
	users, err := db.store.FindRecursive("/users", "name", name)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user by name: %w", err)
	}

	for _, searchResult := range users {
		if user, ok := searchResult.Value.(*core.User); ok && user.Name == name && user.ColonyName == colonyName {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user with name %s not found in colony %s", name, colonyName)
}

// RemoveUserByID removes a user by colony name and user ID
func (db *KVStoreDatabase) RemoveUserByID(colonyName string, userID string) error {
	userPath := fmt.Sprintf("/users/%s", userID)
	
	if !db.store.Exists(userPath) {
		return fmt.Errorf("user with ID %s not found in colony %s", userID, colonyName)
	}

	err := db.store.Delete(userPath)
	if err != nil {
		return fmt.Errorf("failed to remove user %s: %w", userID, err)
	}

	return nil
}

// RemoveUserByName removes a user by colony name and user name
func (db *KVStoreDatabase) RemoveUserByName(colonyName string, name string) error {
	// First find the user to get their ID
	user, err := db.GetUserByName(colonyName, name)
	if err != nil {
		return err
	}

	return db.RemoveUserByID(colonyName, user.ID)
}

// RemoveUsersByColonyName removes all users from a colony
func (db *KVStoreDatabase) RemoveUsersByColonyName(colonyName string) error {
	// Find all users for the colony
	users, err := db.store.FindRecursive("/users", "colonyname", colonyName)
	if err != nil {
		return nil // No users found to remove
	}

	// Remove each user
	for _, searchResult := range users {
		if user, ok := searchResult.Value.(*core.User); ok {
			userPath := fmt.Sprintf("/users/%s", user.ID)
			err := db.store.Delete(userPath)
			if err != nil {
				return fmt.Errorf("failed to remove user %s: %w", user.ID, err)
			}
		}
	}

	return nil
}

// GetUsers retrieves all users from all colonies
func (db *KVStoreDatabase) GetUsers() ([]*core.User, error) {
	// Find all users in the users directory
	users, err := db.store.FindAllRecursive("/users", "userid")
	if err != nil {
		return []*core.User{}, nil // Return empty slice when no users found
	}

	var result []*core.User
	for _, searchResult := range users {
		if user, ok := searchResult.Value.(*core.User); ok {
			result = append(result, user)
		}
	}

	return result, nil
}

// CountUsers returns the total number of users
func (db *KVStoreDatabase) CountUsers() (int, error) {
	users, err := db.GetUsers()
	if err != nil {
		return 0, err
	}
	return len(users), nil
}

// CountUsersByColonyName returns the number of users in a specific colony
func (db *KVStoreDatabase) CountUsersByColonyName(colonyName string) (int, error) {
	users, err := db.GetUsersByColonyName(colonyName)
	if err != nil {
		return 0, err
	}
	return len(users), nil
}