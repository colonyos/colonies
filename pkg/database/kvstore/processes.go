package kvstore

import (
	"errors"
	"fmt"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// ProcessDatabase Interface Implementation
// =====================================

// AddProcess adds a process to the database
func (db *KVStoreDatabase) AddProcess(process *core.Process) error {
	if process == nil {
		return errors.New("process cannot be nil")
	}

	// Store process directly
	processPath := fmt.Sprintf("/processes/%s", process.ID)
	
	// Check if process already exists
	if db.store.Exists(processPath) {
		return fmt.Errorf("process with ID %s already exists", process.ID)
	}

	// Store the process
	err := db.store.Put(processPath, process)
	if err != nil {
		return fmt.Errorf("failed to add process %s: %w", process.ID, err)
	}

	return nil
}

// GetProcesses retrieves all processes
func (db *KVStoreDatabase) GetProcesses() ([]*core.Process, error) {
	// Find all processes (using correct JSON field name "processid")
	processes, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		// Return empty slice when no processes found, like PostgreSQL
		return []*core.Process{}, nil
	}

	var result []*core.Process
	for _, searchResult := range processes {
		if process, ok := searchResult.Value.(*core.Process); ok {
			result = append(result, process)
		}
	}

	return result, nil
}

// GetProcessByID retrieves a process by ID
func (db *KVStoreDatabase) GetProcessByID(processID string) (*core.Process, error) {
	processPath := fmt.Sprintf("/processes/%s", processID)
	
	if !db.store.Exists(processPath) {
		return nil, fmt.Errorf("process with ID %s not found", processID)
	}

	processInterface, err := db.store.Get(processPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get process %s: %w", processID, err)
	}

	storedProcess, ok := processInterface.(*core.Process)
	if !ok {
		return nil, fmt.Errorf("stored object is not a process")
	}

	// Create a copy to prevent race conditions when multiple goroutines access the same object
	processCopy := *storedProcess
	
	// Populate the process's attributes (called without holding the lock to avoid deadlock)
	attributes, err := db.GetAttributes(processID)
	if err != nil {
		// If getting attributes fails, just log it but don't fail the whole process retrieval
		// This maintains compatibility with PostgreSQL behavior
		attributes = []core.Attribute{}
	}
	processCopy.Attributes = attributes

	return &processCopy, nil
}

// FindProcessesByColonyName finds processes by colony name and state within time window
func (db *KVStoreDatabase) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	// Get all processes first, then filter by colony name
	allProcesses, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		// Return empty slice when no processes found, like PostgreSQL
		return []*core.Process{}, nil
	}

	var result []*core.Process
	currentTime := time.Now()
	for _, searchResult := range allProcesses {
		if process, ok := searchResult.Value.(*core.Process); ok {
			// Check colony name match (nested in FunctionSpec.Conditions.ColonyName)
			if process.FunctionSpec.Conditions.ColonyName != colonyName {
				continue
			}
			
			// Check state match
			if state >= 0 && process.State != state {
				continue
			}
			
			// Check time window if specified
			if seconds > 0 {
				timeDiff := currentTime.Sub(process.SubmissionTime)
				if timeDiff.Seconds() > float64(seconds) {
					continue
				}
			}
			
			result = append(result, process)
		}
	}

	return result, nil
}

// FindProcessesByExecutorID finds processes by executor ID within time window
func (db *KVStoreDatabase) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	// Get all processes first, then filter
	allProcesses, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		// Return empty slice when no processes found, like PostgreSQL
		return []*core.Process{}, nil
	}

	var result []*core.Process
	currentTime := time.Now()
	for _, searchResult := range allProcesses {
		if process, ok := searchResult.Value.(*core.Process); ok {
			// Check colony name match
			if process.FunctionSpec.Conditions.ColonyName != colonyName {
				continue
			}
			
			// Check executor ID match
			if process.AssignedExecutorID != executorID {
				continue
			}
			
			// Check state match
			if state >= 0 && process.State != state {
				continue
			}
			
			// Check time window if specified
			if seconds > 0 {
				timeDiff := currentTime.Sub(process.SubmissionTime)
				if timeDiff.Seconds() > float64(seconds) {
					continue
				}
			}
			
			result = append(result, process)
		}
	}

	return result, nil
}

// SetProcessState sets the state of a process
func (db *KVStoreDatabase) SetProcessState(processID string, state int) error {
	processPath := fmt.Sprintf("/processes/%s", processID)
	
	if !db.store.Exists(processPath) {
		return fmt.Errorf("process with ID %s not found", processID)
	}

	processInterface, err := db.store.Get(processPath)
	if err != nil {
		return fmt.Errorf("failed to get process %s: %w", processID, err)
	}

	storedProcess, ok := processInterface.(*core.Process)
	if !ok {
		return fmt.Errorf("stored object is not a process")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	process := *storedProcess
	process.State = state

	// Store back
	err = db.store.Put(processPath, &process)
	if err != nil {
		return fmt.Errorf("failed to update process state: %w", err)
	}

	return nil
}

// RemoveProcessByID removes a process by ID
func (db *KVStoreDatabase) RemoveProcessByID(processID string) error {
	processPath := fmt.Sprintf("/processes/%s", processID)
	
	if !db.store.Exists(processPath) {
		return fmt.Errorf("process with ID %s not found", processID)
	}

	err := db.store.Delete(processPath)
	if err != nil {
		return fmt.Errorf("failed to remove process %s: %w", processID, err)
	}

	return nil
}

// RemoveAllProcesses removes all processes
func (db *KVStoreDatabase) RemoveAllProcesses() error {
	// Get all processes to remove them individually
	processes, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		// If no processes found, that's fine
		return nil
	}

	// Remove each process
	for _, searchResult := range processes {
		if process, ok := searchResult.Value.(*core.Process); ok {
			err := db.RemoveProcessByID(process.ID)
			if err != nil {
				// Continue removing others even if one fails
				continue
			}
		}
	}

	return nil
}

// CountProcesses returns the total number of processes
func (db *KVStoreDatabase) CountProcesses() (int, error) {
	// Find all processes directly (avoid nested locking)
	processes, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		return 0, nil
	}

	return len(processes), nil
}

// FindWaitingProcesses finds waiting processes matching criteria
func (db *KVStoreDatabase) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByStateAndFilters(colonyName, 0, executorType, label, initiator, count) // 0 = WAITING
}

// FindRunningProcesses finds running processes matching criteria
func (db *KVStoreDatabase) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByStateAndFilters(colonyName, 1, executorType, label, initiator, count) // 1 = RUNNING
}

// FindSuccessfulProcesses finds successful processes matching criteria
func (db *KVStoreDatabase) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByStateAndFilters(colonyName, 2, executorType, label, initiator, count) // 2 = SUCCESS
}

// FindFailedProcesses finds failed processes matching criteria
func (db *KVStoreDatabase) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return db.findProcessesByStateAndFilters(colonyName, 3, executorType, label, initiator, count) // 3 = FAILED
}

// Helper method to find processes by state and filters
func (db *KVStoreDatabase) findProcessesByStateAndFilters(colonyName string, state int, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	allProcesses, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		return []*core.Process{}, nil
	}

	var result []*core.Process
	for _, searchResult := range allProcesses {
		if process, ok := searchResult.Value.(*core.Process); ok {
			// Check state
			if process.State != state {
				continue
			}
			
			// Check colony name
			if colonyName != "" && process.FunctionSpec.Conditions.ColonyName != colonyName {
				continue
			}
			
			// Check executor type
			if executorType != "" && process.FunctionSpec.Conditions.ExecutorType != executorType {
				continue
			}
			
			// Check label (this might need adjustment based on how labels are stored)
			if label != "" {
				// Labels might be stored differently - this is a placeholder
				continue
			}
			
			// Check initiator
			if initiator != "" && process.InitiatorName != initiator {
				continue
			}
			
			result = append(result, process)
			
			// Limit results if count specified
			if count > 0 && len(result) >= count {
				break
			}
		}
	}

	return result, nil
}

// Additional process-related methods...

// FindAllRunningProcesses finds all running processes
func (db *KVStoreDatabase) FindAllRunningProcesses() ([]*core.Process, error) {
	return db.findProcessesByState(1) // Assuming 1 = RUNNING
}

// FindAllWaitingProcesses finds all waiting processes
func (db *KVStoreDatabase) FindAllWaitingProcesses() ([]*core.Process, error) {
	return db.findProcessesByState(0) // Assuming 0 = WAITING
}

func (db *KVStoreDatabase) findProcessesByState(state int) ([]*core.Process, error) {
	processes, err := db.GetProcesses()
	if err != nil {
		return nil, err
	}

	var result []*core.Process
	for _, process := range processes {
		if process.State == state {
			result = append(result, process)
		}
	}

	return result, nil
}

// Process field update methods

// SetInput sets the input of a process
func (db *KVStoreDatabase) SetInput(processID string, input []interface{}) error {
	return db.updateProcessField(processID, "input", input)
}

// SetOutput sets the output of a process
func (db *KVStoreDatabase) SetOutput(processID string, output []interface{}) error {
	return db.updateProcessField(processID, "output", output)
}

// SetErrors sets the errors of a process
func (db *KVStoreDatabase) SetErrors(processID string, errs []string) error {
	return db.updateProcessField(processID, "errors", errs)
}

// SetParents sets the parents of a process
func (db *KVStoreDatabase) SetParents(processID string, parents []string) error {
	return db.updateProcessField(processID, "parents", parents)
}

// SetChildren sets the children of a process
func (db *KVStoreDatabase) SetChildren(processID string, children []string) error {
	return db.updateProcessField(processID, "children", children)
}

// SetWaitingForParents sets whether a process is waiting for parents
func (db *KVStoreDatabase) SetWaitingForParents(processID string, waitingForParent bool) error {
	return db.updateProcessField(processID, "waitforparents", waitingForParent)
}

func (db *KVStoreDatabase) updateProcessField(processID string, fieldName string, value interface{}) error {
	processPath := fmt.Sprintf("/processes/%s", processID)
	
	if !db.store.Exists(processPath) {
		return fmt.Errorf("process with ID %s not found", processID)
	}

	processInterface, err := db.store.Get(processPath)
	if err != nil {
		return fmt.Errorf("failed to get process %s: %w", processID, err)
	}

	storedProcess, ok := processInterface.(*core.Process)
	if !ok {
		return fmt.Errorf("stored object is not a process")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	process := *storedProcess

	// Update the field using reflection or direct assignment
	switch fieldName {
	case "input":
		if input, ok := value.([]interface{}); ok {
			process.FunctionSpec.Args = input
		}
	case "output":
		if output, ok := value.([]interface{}); ok {
			process.Output = output
		}
	case "errors":
		if errors, ok := value.([]string); ok {
			process.Errors = errors
		}
	case "parents":
		if parents, ok := value.([]string); ok {
			process.Parents = parents
		}
	case "children":
		if children, ok := value.([]string); ok {
			process.Children = children
		}
	case "waitforparents":
		if waitForParents, ok := value.(bool); ok {
			process.WaitForParents = waitForParents
		}
	}

	// Store back
	err = db.store.Put(processPath, &process)
	if err != nil {
		return fmt.Errorf("failed to update process field %s: %w", fieldName, err)
	}

	return nil
}

// Process assignment methods

// Assign assigns a process to an executor
func (db *KVStoreDatabase) Assign(executorID string, process *core.Process) error {
	if process == nil {
		return errors.New("process cannot be nil")
	}

	// Check if process exists in database first
	processPath := fmt.Sprintf("/processes/%s", process.ID)
	if !db.store.Exists(processPath) {
		return fmt.Errorf("process with ID %s not found", process.ID)
	}

	// Get the current process from storage to avoid race conditions
	storedProcessInterface, err := db.store.Get(processPath)
	if err != nil {
		return fmt.Errorf("failed to get process for assignment: %w", err)
	}

	storedProcess, ok := storedProcessInterface.(*core.Process)
	if !ok {
		return fmt.Errorf("stored object is not a process")
	}

	// Create a copy to avoid modifying the original
	assignedProcess := *storedProcess
	
	// Set executor assignment on the copy
	assignedProcess.AssignedExecutorID = executorID
	assignedProcess.IsAssigned = true
	assignedProcess.State = 1 // RUNNING
	err = db.store.Put(processPath, &assignedProcess)
	if err != nil {
		return fmt.Errorf("failed to assign process %s to executor %s: %w", process.ID, executorID, err)
	}

	return nil
}

// Unassign unassigns a process from its executor
func (db *KVStoreDatabase) Unassign(process *core.Process) error {
	if process == nil {
		return errors.New("process cannot be nil")
	}

	// Check if process exists in database first
	processPath := fmt.Sprintf("/processes/%s", process.ID)
	if !db.store.Exists(processPath) {
		return fmt.Errorf("process with ID %s not found", process.ID)
	}

	// Get the current process from storage to avoid race conditions
	storedProcessInterface, err := db.store.Get(processPath)
	if err != nil {
		return fmt.Errorf("failed to get process for unassignment: %w", err)
	}

	storedProcess, ok := storedProcessInterface.(*core.Process)
	if !ok {
		return fmt.Errorf("stored object is not a process")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	unassignedProcess := *storedProcess
	
	// Clear executor assignment
	unassignedProcess.AssignedExecutorID = ""
	unassignedProcess.IsAssigned = false
	unassignedProcess.State = 0 // WAITING

	err = db.store.Put(processPath, &unassignedProcess)
	if err != nil {
		return fmt.Errorf("failed to unassign process %s: %w", process.ID, err)
	}

	return nil
}

// Process completion methods
func (db *KVStoreDatabase) MarkSuccessful(processID string) (float64, float64, error) {
	err := db.SetProcessState(processID, 2) // SUCCESS
	if err != nil {
		return 0, 0, err
	}
	// Return dummy CPU and memory values - KVStore doesn't track these
	return 0.0, 0.0, nil
}

func (db *KVStoreDatabase) MarkFailed(processID string, errs []string) error {
	process, err := db.GetProcessByID(processID)
	if err != nil {
		return err
	}
	if process == nil {
		return fmt.Errorf("process with ID %s not found", processID)
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedProcess := *process
	updatedProcess.State = core.FAILED
	updatedProcess.Errors = errs
	updatedProcess.EndTime = time.Now()

	// Store back
	processPath := fmt.Sprintf("/processes/%s", processID)
	err = db.store.Put(processPath, &updatedProcess)
	if err != nil {
		return fmt.Errorf("failed to mark process as failed: %w", err)
	}

	return nil
}

// Additional placeholder methods for complex operations
// These would need more sophisticated implementation in a production system

// Find candidates - simplified implementation
func (db *KVStoreDatabase) FindCandidates(colonyName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return db.findProcessesByStateAndFilters(colonyName, 0, executorType, "", "", count) // Return waiting processes
}

func (db *KVStoreDatabase) FindCandidatesByName(colonyName string, executorName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return db.findProcessesByStateAndFilters(colonyName, 0, executorType, "", "", count) // Return waiting processes
}

// Remove processes by state and colony
func (db *KVStoreDatabase) RemoveAllWaitingProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyAndState(colonyName, 0) // WAITING
}

func (db *KVStoreDatabase) RemoveAllRunningProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyAndState(colonyName, 1) // RUNNING
}

func (db *KVStoreDatabase) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyAndState(colonyName, 2) // SUCCESS
}

func (db *KVStoreDatabase) RemoveAllFailedProcessesByColonyName(colonyName string) error {
	return db.removeProcessesByColonyAndState(colonyName, 3) // FAILED
}

// Helper method to remove processes by colony and state
func (db *KVStoreDatabase) removeProcessesByColonyAndState(colonyName string, state int) error {
	processes, err := db.FindProcessesByColonyName(colonyName, 0, state)
	if err != nil {
		return err
	}

	for _, process := range processes {
		err := db.RemoveProcessByID(process.ID)
		if err != nil {
			// Continue removing others even if one fails
			continue
		}
	}

	return nil
}

// RemoveAllProcessesByColonyName removes all processes for a colony
func (db *KVStoreDatabase) RemoveAllProcessesByColonyName(colonyName string) error {
	processes, err := db.FindProcessesByColonyName(colonyName, -1, -1) // Get all processes regardless of state
	if err != nil {
		return err
	}

	for _, process := range processes {
		err := db.RemoveProcessByID(process.ID)
		if err != nil {
			// Continue removing others even if one fails
			continue
		}
	}

	return nil
}

// RemoveAllProcessesByProcessGraphID removes all processes for a process graph
func (db *KVStoreDatabase) RemoveAllProcessesByProcessGraphID(processGraphID string) error {
	// Find all processes with this process graph ID
	allProcesses, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		return nil
	}

	for _, searchResult := range allProcesses {
		if process, ok := searchResult.Value.(*core.Process); ok {
			if process.ProcessGraphID == processGraphID {
				err := db.RemoveProcessByID(process.ID)
				if err != nil {
					// Continue removing others even if one fails
					continue
				}
			}
		}
	}

	return nil
}

// ResetProcess resets a process to waiting state
func (db *KVStoreDatabase) ResetProcess(process *core.Process) error {
	if process == nil {
		return errors.New("process cannot be nil")
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedProcess := *process
	updatedProcess.State = core.WAITING
	updatedProcess.IsAssigned = false
	updatedProcess.AssignedExecutorID = ""
	updatedProcess.StartTime = time.Time{}
	updatedProcess.EndTime = time.Time{}
	updatedProcess.SubmissionTime = time.Now()
	updatedProcess.Errors = []string{}

	// Set wait deadline if MaxWaitTime is specified
	if process.FunctionSpec.MaxWaitTime > 0 {
		updatedProcess.WaitDeadline = time.Now().Add(time.Duration(process.FunctionSpec.MaxWaitTime) * time.Second)
	}

	// Store back
	processPath := fmt.Sprintf("/processes/%s", process.ID)
	err := db.store.Put(processPath, &updatedProcess)
	if err != nil {
		return fmt.Errorf("failed to reset process: %w", err)
	}

	return nil
}

// SetWaitForParents sets the wait for parents flag for a process
func (db *KVStoreDatabase) SetWaitForParents(processID string, waitForParent bool) error {
	process, err := db.GetProcessByID(processID)
	if err != nil {
		return err
	}

	// Create a copy to avoid modifying the original (race condition fix)
	updatedProcess := *process
	updatedProcess.WaitForParents = waitForParent

	// Store back
	processPath := fmt.Sprintf("/processes/%s", processID)
	err = db.store.Put(processPath, &updatedProcess)
	if err != nil {
		return fmt.Errorf("failed to set wait for parents: %w", err)
	}

	return nil
}

// RemoveAllProcessesInProcessGraphsByColonyName removes all processes in process graphs for a colony
func (db *KVStoreDatabase) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error {
	// This would ideally find all process graphs for the colony and remove processes in them
	// For now, implement as removing all processes by colony name
	return db.RemoveAllProcessesByColonyName(colonyName)
}

// RemoveAllProcessesInProcessGraphsByColonyNameWithState removes all processes in process graphs for a colony with state
func (db *KVStoreDatabase) RemoveAllProcessesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	// Find all processes in the colony with the specified state
	processes, err := db.FindProcessesByColonyName(colonyName, -1, state)
	if err != nil {
		return err
	}

	for _, process := range processes {
		// Only remove processes that are part of a process graph
		if process.ProcessGraphID != "" {
			err := db.RemoveProcessByID(process.ID)
			if err != nil {
				// Continue removing others even if one fails
				continue
			}
		}
	}

	return nil
}

// CountFailedProcesses counts all failed processes
func (db *KVStoreDatabase) CountFailedProcesses() (int, error) {
	return db.countProcessesByState(core.FAILED)
}

// CountRunningProcesses counts all running processes
func (db *KVStoreDatabase) CountRunningProcesses() (int, error) {
	return db.countProcessesByState(core.RUNNING)
}

// CountWaitingProcesses counts all waiting processes
func (db *KVStoreDatabase) CountWaitingProcesses() (int, error) {
	return db.countProcessesByState(core.WAITING)
}

// CountSuccessfulProcesses counts all successful processes
func (db *KVStoreDatabase) CountSuccessfulProcesses() (int, error) {
	return db.countProcessesByState(core.SUCCESS)
}

// CountFailedProcessesByColonyName counts failed processes by colony name
func (db *KVStoreDatabase) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	processes, err := db.FindFailedProcesses(colonyName, "", "", "", -1)
	if err != nil {
		return 0, err
	}

	return len(processes), nil
}

// CountRunningProcessesByColonyName counts running processes by colony name
func (db *KVStoreDatabase) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	processes, err := db.FindRunningProcesses(colonyName, "", "", "", -1)
	if err != nil {
		return 0, err
	}

	return len(processes), nil
}

// CountWaitingProcessesByColonyName counts waiting processes by colony name
func (db *KVStoreDatabase) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	processes, err := db.FindWaitingProcesses(colonyName, "", "", "", -1)
	if err != nil {
		return 0, err
	}

	return len(processes), nil
}

// CountSuccessfulProcessesByColonyName counts successful processes by colony name
func (db *KVStoreDatabase) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	processes, err := db.FindSuccessfulProcesses(colonyName, "", "", "", -1)
	if err != nil {
		return 0, err
	}

	return len(processes), nil
}

// Helper method to count processes by state
func (db *KVStoreDatabase) countProcessesByState(state int) (int, error) {
	// Get all processes first, then count those with matching state
	allProcesses, err := db.store.FindAllRecursive("/processes", "processid")
	if err != nil {
		return 0, nil
	}

	count := 0
	for _, searchResult := range allProcesses {
		if process, ok := searchResult.Value.(*core.Process); ok {
			if process.State == state {
				count++
			}
		}
	}

	return count, nil
}