package kvstore

import (
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// FunctionDatabase Interface Implementation
// =====================================

// AddFunction adds a function to the database
func (db *KVStoreDatabase) AddFunction(function *core.Function) error {
	if function == nil {
		return errors.New("function cannot be nil")
	}

	// Store function at /functions/{functionID} (upsert behavior)
	functionPath := fmt.Sprintf("/functions/%s", function.FunctionID)
	
	err := db.store.Put(functionPath, function)
	if err != nil {
		return fmt.Errorf("failed to add function %s: %w", function.FunctionID, err)
	}

	return nil
}

// GetFunctionByID retrieves a function by ID
func (db *KVStoreDatabase) GetFunctionByID(functionID string) (*core.Function, error) {
	functionPath := fmt.Sprintf("/functions/%s", functionID)
	
	if !db.store.Exists(functionPath) {
		// Return (nil, nil) when function not found, like PostgreSQL
		return nil, nil
	}

	functionInterface, err := db.store.Get(functionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get function %s: %w", functionID, err)
	}

	function, ok := functionInterface.(*core.Function)
	if !ok {
		return nil, fmt.Errorf("stored object is not a function")
	}

	return function, nil
}

// GetFunctionsByExecutorName retrieves functions by colony name and executor name
func (db *KVStoreDatabase) GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error) {
	// Search for functions by colony name first, then filter by executor name
	functions, err := db.store.FindRecursive("/functions", "colonyname", colonyName)
	if err != nil {
		return []*core.Function{}, nil
	}

	var result []*core.Function
	for _, searchResult := range functions {
		if function, ok := searchResult.Value.(*core.Function); ok {
			if function.ExecutorName == executorName {
				result = append(result, function)
			}
		}
	}

	return result, nil
}

// GetFunctionsByColonyName retrieves functions by colony name
func (db *KVStoreDatabase) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	// Search for functions by colony name
	functions, err := db.store.FindRecursive("/functions", "colonyname", colonyName)
	if err != nil {
		return []*core.Function{}, nil
	}

	var result []*core.Function
	for _, searchResult := range functions {
		if function, ok := searchResult.Value.(*core.Function); ok {
			result = append(result, function)
		}
	}

	return result, nil
}

// GetFunctionsByExecutorAndName retrieves a function by colony name, executor name, and function name
func (db *KVStoreDatabase) GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error) {
	// Search for functions by colony name first
	functions, err := db.store.FindRecursive("/functions", "colonyname", colonyName)
	if err != nil {
		return nil, nil
	}

	for _, searchResult := range functions {
		if function, ok := searchResult.Value.(*core.Function); ok {
			if function.ExecutorName == executorName && function.FuncName == name {
				return function, nil
			}
		}
	}

	return nil, nil
}

// UpdateFunctionStats updates function statistics
func (db *KVStoreDatabase) UpdateFunctionStats(colonyName string, executorName string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error {
	// Find the function
	function, err := db.GetFunctionsByExecutorAndName(colonyName, executorName, name)
	if err != nil {
		return err
	}
	if function == nil {
		return fmt.Errorf("function with name %s not found for executor %s in colony %s", name, executorName, colonyName)
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedFunction := *function
	updatedFunction.Counter = counter
	updatedFunction.MinWaitTime = minWaitTime
	updatedFunction.MaxWaitTime = maxWaitTime
	updatedFunction.MinExecTime = minExecTime
	updatedFunction.MaxExecTime = maxExecTime
	updatedFunction.AvgWaitTime = avgWaitTime
	updatedFunction.AvgExecTime = avgExecTime

	// Store back
	functionPath := fmt.Sprintf("/functions/%s", function.FunctionID)
	err = db.store.Put(functionPath, &updatedFunction)
	if err != nil {
		return fmt.Errorf("failed to update function stats: %w", err)
	}

	return nil
}

// RemoveFunctionByID removes a function by ID
func (db *KVStoreDatabase) RemoveFunctionByID(functionID string) error {
	functionPath := fmt.Sprintf("/functions/%s", functionID)
	
	if !db.store.Exists(functionPath) {
		return fmt.Errorf("function with ID %s not found", functionID)
	}

	err := db.store.Delete(functionPath)
	if err != nil {
		return fmt.Errorf("failed to remove function %s: %w", functionID, err)
	}

	return nil
}

// RemoveFunctionByName removes a function by colony name, executor name, and function name
func (db *KVStoreDatabase) RemoveFunctionByName(colonyName string, executorName string, name string) error {
	function, err := db.GetFunctionsByExecutorAndName(colonyName, executorName, name)
	if err != nil {
		return err
	}
	if function == nil {
		return fmt.Errorf("function with name %s not found for executor %s in colony %s", name, executorName, colonyName)
	}

	return db.RemoveFunctionByID(function.FunctionID)
}

// RemoveFunctionsByExecutorName removes all functions for an executor
func (db *KVStoreDatabase) RemoveFunctionsByExecutorName(colonyName string, executorName string) error {
	functions, err := db.GetFunctionsByExecutorName(colonyName, executorName)
	if err != nil {
		return err
	}

	for _, function := range functions {
		err := db.RemoveFunctionByID(function.FunctionID)
		if err != nil {
			return fmt.Errorf("failed to remove function %s: %w", function.FunctionID, err)
		}
	}

	return nil
}

// RemoveFunctionsByColonyName removes all functions for a colony
func (db *KVStoreDatabase) RemoveFunctionsByColonyName(colonyName string) error {
	functions, err := db.GetFunctionsByColonyName(colonyName)
	if err != nil {
		return err
	}

	for _, function := range functions {
		err := db.RemoveFunctionByID(function.FunctionID)
		if err != nil {
			return fmt.Errorf("failed to remove function %s: %w", function.FunctionID, err)
		}
	}

	return nil
}

// RemoveFunctions removes all functions from the database
func (db *KVStoreDatabase) RemoveFunctions() error {
	functionsPath := "/functions"
	
	if !db.store.Exists(functionsPath) {
		return nil // No functions to remove
	}

	err := db.store.Delete(functionsPath)
	if err != nil {
		return fmt.Errorf("failed to remove all functions: %w", err)
	}

	return nil
}