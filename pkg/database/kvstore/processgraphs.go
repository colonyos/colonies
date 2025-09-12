package kvstore

import (
	"errors"
	"fmt"

	"github.com/colonyos/colonies/pkg/core"
)

// ProcessGraphDatabase Interface Implementation
// =====================================

// AddProcessGraph adds a process graph to the database
func (db *KVStoreDatabase) AddProcessGraph(processGraph *core.ProcessGraph) error {
	if processGraph == nil {
		return errors.New("process graph cannot be nil")
	}

	// Store process graph at /processgraphs/{processGraphID}
	processGraphPath := fmt.Sprintf("/processgraphs/%s", processGraph.ID)
	
	// Check if process graph already exists
	if db.store.Exists(processGraphPath) {
		return fmt.Errorf("process graph with ID %s already exists", processGraph.ID)
	}

	err := db.store.Put(processGraphPath, processGraph)
	if err != nil {
		return fmt.Errorf("failed to add process graph %s: %w", processGraph.ID, err)
	}

	return nil
}

// GetProcessGraphByID retrieves a process graph by ID
func (db *KVStoreDatabase) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	// Handle empty/null IDs gracefully - return (nil, nil) like PostgreSQL
	if processGraphID == "" {
		return nil, nil
	}

	processGraphPath := fmt.Sprintf("/processgraphs/%s", processGraphID)
	
	if !db.store.Exists(processGraphPath) {
		// Return (nil, nil) when process graph not found, like PostgreSQL
		return nil, nil
	}

	processGraphInterface, err := db.store.Get(processGraphPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get process graph %s: %w", processGraphID, err)
	}

	storedProcessGraph, ok := processGraphInterface.(*core.ProcessGraph)
	if !ok {
		return nil, fmt.Errorf("stored object is not a process graph")
	}

	// Return a copy to prevent race conditions when multiple goroutines access the same object
	processGraphCopy := *storedProcessGraph
	return &processGraphCopy, nil
}

// SetProcessGraphState sets the state of a process graph
func (db *KVStoreDatabase) SetProcessGraphState(processGraphID string, state int) error {
	// Handle empty IDs
	if processGraphID == "" {
		return fmt.Errorf("process graph ID cannot be empty")
	}

	processGraphPath := fmt.Sprintf("/processgraphs/%s", processGraphID)
	
	if !db.store.Exists(processGraphPath) {
		return fmt.Errorf("process graph with ID %s not found", processGraphID)
	}

	processGraphInterface, err := db.store.Get(processGraphPath)
	if err != nil {
		return fmt.Errorf("failed to get process graph %s: %w", processGraphID, err)
	}

	storedProcessGraph, ok := processGraphInterface.(*core.ProcessGraph)
	if !ok {
		return fmt.Errorf("stored object is not a process graph")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	processGraph := *storedProcessGraph
	processGraph.State = state

	// Store back
	err = db.store.Put(processGraphPath, &processGraph)
	if err != nil {
		return fmt.Errorf("failed to update process graph state: %w", err)
	}

	return nil
}

// RemoveProcessGraphByID removes a process graph by ID
func (db *KVStoreDatabase) RemoveProcessGraphByID(processGraphID string) error {
	// Handle empty IDs
	if processGraphID == "" {
		return fmt.Errorf("process graph ID cannot be empty")
	}

	processGraphPath := fmt.Sprintf("/processgraphs/%s", processGraphID)
	
	if !db.store.Exists(processGraphPath) {
		return fmt.Errorf("process graph with ID %s not found", processGraphID)
	}

	err := db.store.Delete(processGraphPath)
	if err != nil {
		return fmt.Errorf("failed to remove process graph %s: %w", processGraphID, err)
	}

	return nil
}

// FindProcessGraphsByColonyName finds process graphs by colony name and state
func (db *KVStoreDatabase) FindProcessGraphsByColonyName(colonyName string, state int) ([]*core.ProcessGraph, error) {
	// Get all process graphs first, then filter by colony name and state
	allProcessGraphs, err := db.store.FindAllRecursive("/processgraphs", "processgraphid")
	if err != nil {
		// Return empty slice when no process graphs found, like PostgreSQL
		return []*core.ProcessGraph{}, nil
	}

	var result []*core.ProcessGraph
	for _, searchResult := range allProcessGraphs {
		if processGraph, ok := searchResult.Value.(*core.ProcessGraph); ok {
			// Check colony name match
			if processGraph.ColonyName != colonyName {
				continue
			}
			
			// Check state match
			if state >= 0 && processGraph.State != state {
				continue
			}
			
			result = append(result, processGraph)
		}
	}

	return result, nil
}

// CountProcessGraphsByColonyName counts process graphs by colony name and state
func (db *KVStoreDatabase) CountProcessGraphsByColonyName(colonyName string, state int) (int, error) {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, state)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, processGraph := range processGraphs {
		if processGraph.State == state {
			count++
		}
	}

	return count, nil
}

// RemoveAllProcessGraphsByColonyName removes all process graphs for a colony
func (db *KVStoreDatabase) RemoveAllProcessGraphsByColonyName(colonyName string) error {
	// Get all process graphs first, then remove those belonging to the colony
	allProcessGraphs, err := db.store.FindAllRecursive("/processgraphs", "processgraphid")
	if err != nil {
		// No process graphs found to remove, that's okay
		return nil
	}

	for _, searchResult := range allProcessGraphs {
		if processGraph, ok := searchResult.Value.(*core.ProcessGraph); ok {
			if processGraph.ColonyName == colonyName {
				err := db.RemoveProcessGraphByID(processGraph.ID)
				if err != nil {
					// Continue removing others even if one fails
					continue
				}
			}
		}
	}

	return nil
}

// RemoveAllFailedProcessGraphsByColonyName removes all failed process graphs for a colony
func (db *KVStoreDatabase) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.FAILED)
	if err != nil {
		return err
	}

	for _, processGraph := range processGraphs {
		err := db.RemoveProcessGraphByID(processGraph.ID)
		if err != nil {
			return fmt.Errorf("failed to remove process graph %s: %w", processGraph.ID, err)
		}
	}

	return nil
}

// RemoveAllRunningProcessGraphsByColonyName removes all running process graphs for a colony
func (db *KVStoreDatabase) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.RUNNING)
	if err != nil {
		return err
	}

	for _, processGraph := range processGraphs {
		err := db.RemoveProcessGraphByID(processGraph.ID)
		if err != nil {
			return fmt.Errorf("failed to remove process graph %s: %w", processGraph.ID, err)
		}
	}

	return nil
}

// RemoveAllWaitingProcessGraphsByColonyName removes all waiting process graphs for a colony
func (db *KVStoreDatabase) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.WAITING)
	if err != nil {
		return err
	}

	for _, processGraph := range processGraphs {
		err := db.RemoveProcessGraphByID(processGraph.ID)
		if err != nil {
			return fmt.Errorf("failed to remove process graph %s: %w", processGraph.ID, err)
		}
	}

	return nil
}

// RemoveAllSuccessfulProcessGraphsByColonyName removes all successful process graphs for a colony
func (db *KVStoreDatabase) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.SUCCESS)
	if err != nil {
		return err
	}

	for _, processGraph := range processGraphs {
		err := db.RemoveProcessGraphByID(processGraph.ID)
		if err != nil {
			return fmt.Errorf("failed to remove process graph %s: %w", processGraph.ID, err)
		}
	}

	return nil
}

// Helper methods for ProcessGraphDatabase

func (db *KVStoreDatabase) findProcessGraphsByColonyNameAndState(colonyName string, state int) ([]*core.ProcessGraph, error) {
	// Get all process graphs first, then filter by colony name and state
	allProcessGraphs, err := db.store.FindAllRecursive("/processgraphs", "processgraphid")
	if err != nil {
		// Return empty slice when no process graphs found, like PostgreSQL
		return []*core.ProcessGraph{}, nil
	}

	var result []*core.ProcessGraph
	for _, searchResult := range allProcessGraphs {
		if processGraph, ok := searchResult.Value.(*core.ProcessGraph); ok {
			// Check colony name match
			if processGraph.ColonyName != colonyName {
				continue
			}
			
			// Check state match
			if state >= 0 && processGraph.State != state {
				continue
			}
			
			result = append(result, processGraph)
		}
	}

	return result, nil
}

// FindFailedProcessGraphs finds all failed process graphs
func (db *KVStoreDatabase) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.FindProcessGraphsByColonyName(colonyName, core.FAILED)
}

// FindRunningProcessGraphs finds all running process graphs
func (db *KVStoreDatabase) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.FindProcessGraphsByColonyName(colonyName, core.RUNNING)
}

// FindWaitingProcessGraphs finds all waiting process graphs
func (db *KVStoreDatabase) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.FindProcessGraphsByColonyName(colonyName, core.WAITING)
}

// FindSuccessfulProcessGraphs finds all successful process graphs
func (db *KVStoreDatabase) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return db.FindProcessGraphsByColonyName(colonyName, core.SUCCESS)
}

// CountFailedProcessGraphs counts all failed process graphs
func (db *KVStoreDatabase) CountFailedProcessGraphs() (int, error) {
	return db.countProcessGraphsByState(core.FAILED)
}

// CountRunningProcessGraphs counts all running process graphs
func (db *KVStoreDatabase) CountRunningProcessGraphs() (int, error) {
	return db.countProcessGraphsByState(core.RUNNING)
}

// CountWaitingProcessGraphs counts all waiting process graphs
func (db *KVStoreDatabase) CountWaitingProcessGraphs() (int, error) {
	return db.countProcessGraphsByState(core.WAITING)
}

// CountSuccessfulProcessGraphs counts all successful process graphs
func (db *KVStoreDatabase) CountSuccessfulProcessGraphs() (int, error) {
	return db.countProcessGraphsByState(core.SUCCESS)
}

// CountFailedProcessGraphsByColonyName counts failed process graphs by colony name
func (db *KVStoreDatabase) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.FAILED)
	if err != nil {
		return 0, err
	}

	return len(processGraphs), nil
}

// CountRunningProcessGraphsByColonyName counts running process graphs by colony name
func (db *KVStoreDatabase) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.RUNNING)
	if err != nil {
		return 0, err
	}

	return len(processGraphs), nil
}

// CountWaitingProcessGraphsByColonyName counts waiting process graphs by colony name
func (db *KVStoreDatabase) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.WAITING)
	if err != nil {
		return 0, err
	}

	return len(processGraphs), nil
}

// CountSuccessfulProcessGraphsByColonyName counts successful process graphs by colony name
func (db *KVStoreDatabase) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) {
	processGraphs, err := db.FindProcessGraphsByColonyName(colonyName, core.SUCCESS)
	if err != nil {
		return 0, err
	}

	return len(processGraphs), nil
}

// Helper method to count process graphs by state
func (db *KVStoreDatabase) countProcessGraphsByState(state int) (int, error) {
	// Get all process graphs first, then count those with matching state
	allProcessGraphs, err := db.store.FindAllRecursive("/processgraphs", "processgraphid")
	if err != nil {
		return 0, nil
	}

	count := 0
	for _, searchResult := range allProcessGraphs {
		if processGraph, ok := searchResult.Value.(*core.ProcessGraph); ok {
			if processGraph.State == state {
				count++
			}
		}
	}

	return count, nil
}