package kvstore

import (
	"errors"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// ExecutorDatabase Interface Implementation
// =====================================

// AddExecutor adds an executor to the database
func (db *KVStoreDatabase) AddExecutor(executor *core.Executor) error {
	if executor == nil {
		return errors.New("executor cannot be nil")
	}

	// Check if executor with same name already exists in colony
	existingExecutor, err := db.GetExecutorByName(executor.ColonyName, executor.Name)
	if err == nil && existingExecutor != nil {
		return fmt.Errorf("executor with name %s already exists in colony %s", executor.Name, executor.ColonyName)
	}

	// Store executor at /executors/{executorID}
	executorPath := fmt.Sprintf("/executors/%s", executor.ID)
	
	// Check if executor ID already exists
	if db.store.Exists(executorPath) {
		return fmt.Errorf("executor with ID %s already exists", executor.ID)
	}

	err = db.store.Put(executorPath, executor)
	if err != nil {
		return fmt.Errorf("failed to add executor %s: %w", executor.ID, err)
	}

	return nil
}

// SetAllocations sets allocations for an executor
func (db *KVStoreDatabase) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	executor, err := db.GetExecutorByName(colonyName, executorName)
	if err != nil {
		return err
	}
	if executor == nil {
		return fmt.Errorf("executor with name %s not found in colony %s", executorName, colonyName)
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedExecutor := *executor
	updatedExecutor.Allocations = allocations

	// Store back
	executorPath := fmt.Sprintf("/executors/%s", executor.ID)
	err = db.store.Put(executorPath, &updatedExecutor)
	if err != nil {
		return fmt.Errorf("failed to update executor allocations: %w", err)
	}

	return nil
}

// GetExecutors retrieves all executors
func (db *KVStoreDatabase) GetExecutors() ([]*core.Executor, error) {
	// Find all executors
	executors, err := db.store.FindAllRecursive("/executors", "executorid")
	if err != nil {
		return []*core.Executor{}, nil
	}

	var result []*core.Executor
	for _, searchResult := range executors {
		if executor, ok := searchResult.Value.(*core.Executor); ok {
			result = append(result, executor)
		}
	}

	return result, nil
}

// GetExecutorByID retrieves an executor by ID
func (db *KVStoreDatabase) GetExecutorByID(executorID string) (*core.Executor, error) {
	executorPath := fmt.Sprintf("/executors/%s", executorID)
	
	if !db.store.Exists(executorPath) {
		return nil, fmt.Errorf("executor with ID %s not found", executorID)
	}

	executorInterface, err := db.store.Get(executorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get executor %s: %w", executorID, err)
	}

	executor, ok := executorInterface.(*core.Executor)
	if !ok {
		return nil, fmt.Errorf("stored object is not an executor")
	}

	return executor, nil
}

// GetExecutorsByColonyName retrieves executors by colony name
func (db *KVStoreDatabase) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	// Search for executors by colony name
	executors, err := db.store.FindRecursive("/executors", "colonyname", colonyName)
	if err != nil {
		return []*core.Executor{}, nil
	}

	var result []*core.Executor
	for _, searchResult := range executors {
		if executor, ok := searchResult.Value.(*core.Executor); ok {
			result = append(result, executor)
		}
	}

	return result, nil
}

// GetExecutorByName retrieves an executor by colony name and executor name
func (db *KVStoreDatabase) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	// Search for executor by colony name and then filter by name
	executors, err := db.store.FindRecursive("/executors", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("executor with name %s not found in colony %s", executorName, colonyName)
	}

	for _, searchResult := range executors {
		if executor, ok := searchResult.Value.(*core.Executor); ok {
			if executor.Name == executorName {
				return executor, nil
			}
		}
	}

	return nil, fmt.Errorf("executor with name %s not found in colony %s", executorName, colonyName)
}

// ApproveExecutor approves an executor
func (db *KVStoreDatabase) ApproveExecutor(executor *core.Executor) error {
	if executor == nil {
		return errors.New("executor cannot be nil")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedExecutor := *executor
	updatedExecutor.State = core.APPROVED
	updatedExecutor.CommissionTime = time.Now()

	// Store back
	executorPath := fmt.Sprintf("/executors/%s", executor.ID)
	err := db.store.Put(executorPath, &updatedExecutor)
	if err != nil {
		return fmt.Errorf("failed to approve executor: %w", err)
	}

	return nil
}

// RejectExecutor rejects an executor
func (db *KVStoreDatabase) RejectExecutor(executor *core.Executor) error {
	if executor == nil {
		return errors.New("executor cannot be nil")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedExecutor := *executor
	updatedExecutor.State = core.REJECTED

	// Store back
	executorPath := fmt.Sprintf("/executors/%s", executor.ID)
	err := db.store.Put(executorPath, &updatedExecutor)
	if err != nil {
		return fmt.Errorf("failed to reject executor: %w", err)
	}

	return nil
}

// MarkAlive marks an executor as alive with current timestamp
func (db *KVStoreDatabase) MarkAlive(executor *core.Executor) error {
	if executor == nil {
		return errors.New("executor cannot be nil")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedExecutor := *executor
	updatedExecutor.LastHeardFromTime = time.Now()

	// Store back
	executorPath := fmt.Sprintf("/executors/%s", executor.ID)
	err := db.store.Put(executorPath, &updatedExecutor)
	if err != nil {
		return fmt.Errorf("failed to mark executor as alive: %w", err)
	}

	return nil
}

// RemoveExecutorByName removes an executor by colony name and executor name
func (db *KVStoreDatabase) RemoveExecutorByName(colonyName string, executorName string) error {
	executor, err := db.GetExecutorByName(colonyName, executorName)
	if err != nil {
		return err
	}
	if executor == nil {
		return fmt.Errorf("executor with name %s not found in colony %s", executorName, colonyName)
	}

	executorPath := fmt.Sprintf("/executors/%s", executor.ID)
	err = db.store.Delete(executorPath)
	if err != nil {
		return fmt.Errorf("failed to remove executor %s: %w", executor.ID, err)
	}

	return nil
}

// RemoveExecutorsByColonyName removes all executors for a colony
func (db *KVStoreDatabase) RemoveExecutorsByColonyName(colonyName string) error {
	// Find all executors for the colony
	executors, err := db.store.FindRecursive("/executors", "colonyname", colonyName)
	if err != nil {
		return nil // No executors found to remove
	}

	// Remove each executor
	for _, searchResult := range executors {
		if executor, ok := searchResult.Value.(*core.Executor); ok {
			executorPath := fmt.Sprintf("/executors/%s", executor.ID)
			err := db.store.Delete(executorPath)
			if err != nil {
				return fmt.Errorf("failed to remove executor %s: %w", executor.ID, err)
			}
		}
	}

	return nil
}

// CountExecutors returns the total number of executors
func (db *KVStoreDatabase) CountExecutors() (int, error) {
	executors, err := db.GetExecutors()
	if err != nil {
		return 0, err
	}

	return len(executors), nil
}

// CountExecutorsByColonyName returns the number of executors in a colony
func (db *KVStoreDatabase) CountExecutorsByColonyName(colonyName string) (int, error) {
	executors, err := db.GetExecutorsByColonyName(colonyName)
	if err != nil {
		return 0, err
	}

	return len(executors), nil
}