package kvstore

import (
	"fmt"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

// =====================================
// LogDatabase Interface Implementation
// =====================================

// AddLog adds a log entry to the database
func (db *KVStoreDatabase) AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error {
	// Create a log entry
	log := &core.Log{
		ProcessID:    processID,
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Timestamp:    timestamp,
		Message:      msg,
	}

	// Generate a unique log ID for storage
	logID := fmt.Sprintf("%s_%d_%d", processID, timestamp, time.Now().UnixNano())

	// Store log at /logs/{logID}
	logPath := fmt.Sprintf("/logs/%s", logID)
	
	err := db.store.Put(logPath, log)
	if err != nil {
		return fmt.Errorf("failed to add log %s: %w", logID, err)
	}

	return nil
}

// GetLogsByProcessID retrieves logs by process ID with limit
func (db *KVStoreDatabase) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) {
	// Search for logs by process ID
	logs, err := db.store.FindRecursive("/logs", "processid", processID)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs for process %s: %w", processID, err)
	}

	var result []*core.Log
	for _, searchResult := range logs {
		if log, ok := searchResult.Value.(*core.Log); ok {
			result = append(result, log)
			
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// GetLogsByProcessIDSince retrieves logs by process ID since timestamp with limit
func (db *KVStoreDatabase) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) {
	// Search for logs by process ID
	logs, err := db.store.FindRecursive("/logs", "processid", processID)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs for process %s: %w", processID, err)
	}

	var result []*core.Log
	for _, searchResult := range logs {
		if log, ok := searchResult.Value.(*core.Log); ok {
			if log.Timestamp > since {
				result = append(result, log)
				
				if limit > 0 && len(result) >= limit {
					break
				}
			}
		}
	}

	return result, nil
}

// GetLogsByExecutor retrieves logs by executor name with limit
func (db *KVStoreDatabase) GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error) {
	// Search for logs by executor name
	logs, err := db.store.FindRecursive("/logs", "executorname", executorName)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs for executor %s: %w", executorName, err)
	}

	var result []*core.Log
	for _, searchResult := range logs {
		if log, ok := searchResult.Value.(*core.Log); ok {
			result = append(result, log)
			
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// GetLogsByExecutorSince retrieves logs by executor name since timestamp with limit
func (db *KVStoreDatabase) GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error) {
	// Search for logs by executor name
	logs, err := db.store.FindRecursive("/logs", "executorname", executorName)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs for executor %s: %w", executorName, err)
	}

	var result []*core.Log
	for _, searchResult := range logs {
		if log, ok := searchResult.Value.(*core.Log); ok {
			if log.Timestamp > since {
				result = append(result, log)
				
				if limit > 0 && len(result) >= limit {
					break
				}
			}
		}
	}

	return result, nil
}

// RemoveLogsByColonyName removes all logs for a colony
func (db *KVStoreDatabase) RemoveLogsByColonyName(colonyName string) error {
	// Find all logs for the colony
	logs, err := db.store.FindRecursive("/logs", "colonyname", colonyName)
	if err != nil {
		return fmt.Errorf("failed to find logs for colony %s: %w", colonyName, err)
	}

	// Remove each log
	for _, searchResult := range logs {
		if _, ok := searchResult.Value.(*core.Log); ok {
			// Since Log doesn't have ID field, we need to use search path
			err := db.store.Delete(searchResult.Path)
			if err != nil {
				return fmt.Errorf("failed to remove log at %s: %w", searchResult.Path, err)
			}
		}
	}

	return nil
}

// CountLogs counts logs for a colony
func (db *KVStoreDatabase) CountLogs(colonyName string) (int, error) {
	// Search for logs by colony name
	logs, err := db.store.FindRecursive("/logs", "colonyname", colonyName)
	if err != nil {
		return 0, fmt.Errorf("failed to find logs for colony %s: %w", colonyName, err)
	}

	count := 0
	for _, searchResult := range logs {
		if _, ok := searchResult.Value.(*core.Log); ok {
			count++
		}
	}

	return count, nil
}

// SearchLogs searches logs by text within days with count limit
func (db *KVStoreDatabase) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) {
	// Search for logs by colony name
	logs, err := db.store.FindRecursive("/logs", "colonyname", colonyName)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs for colony %s: %w", colonyName, err)
	}

	// Calculate time window
	searchSince := time.Now().AddDate(0, 0, -days).Unix()

	var result []*core.Log
	for _, searchResult := range logs {
		if log, ok := searchResult.Value.(*core.Log); ok {
			// Check time window
			if days > 0 && log.Timestamp < searchSince {
				continue
			}
			
			// Check text match
			if text != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(text)) {
				continue
			}
			
			result = append(result, log)
			
			if count > 0 && len(result) >= count {
				break
			}
		}
	}

	return result, nil
}