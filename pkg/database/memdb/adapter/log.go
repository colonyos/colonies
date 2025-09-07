package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/memdb"
)

// LogDatabase interface implementation

func (a *ColonyOSAdapter) AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error {
	log := &core.Log{
		ProcessID:    processID,
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Timestamp:    timestamp,
		Message:      msg,
	}
	
	// Generate a unique ID for the log
	logID := fmt.Sprintf("%s_%d_%d", processID, timestamp, len(msg))
	
	doc := &memdb.VelocityDocument{
		ID:     logID,
		Fields: a.logToFields(log),
	}
	
	return a.db.Insert(context.Background(), LogsCollection, doc)
}

func (a *ColonyOSAdapter) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) {
	result, err := a.db.List(context.Background(), LogsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var logs []*core.Log
	count := 0
	for _, doc := range result {
		if limit > 0 && count >= limit {
			break
		}
		
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ProcessID == processID {
			logs = append(logs, log)
			count++
		}
	}
	
	return logs, nil
}

func (a *ColonyOSAdapter) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) {
	result, err := a.db.List(context.Background(), LogsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var logs []*core.Log
	count := 0
	for _, doc := range result {
		if limit > 0 && count >= limit {
			break
		}
		
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ProcessID == processID && log.Timestamp >= since {
			logs = append(logs, log)
			count++
		}
	}
	
	return logs, nil
}

func (a *ColonyOSAdapter) GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error) {
	result, err := a.db.List(context.Background(), LogsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var logs []*core.Log
	count := 0
	for _, doc := range result {
		if limit > 0 && count >= limit {
			break
		}
		
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ExecutorName == executorName {
			logs = append(logs, log)
			count++
		}
	}
	
	return logs, nil
}

func (a *ColonyOSAdapter) GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error) {
	result, err := a.db.List(context.Background(), LogsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	var logs []*core.Log
	count := 0
	for _, doc := range result {
		if limit > 0 && count >= limit {
			break
		}
		
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ExecutorName == executorName && log.Timestamp >= since {
			logs = append(logs, log)
			count++
		}
	}
	
	return logs, nil
}

func (a *ColonyOSAdapter) CountLogs(colonyName string) (int, error) {
	result, err := a.db.List(context.Background(), LogsCollection, 10000, 0)
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, doc := range result {
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ColonyName == colonyName {
			count++
		}
	}
	
	return count, nil
}

func (a *ColonyOSAdapter) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) {
	result, err := a.db.List(context.Background(), LogsCollection, 1000, 0)
	if err != nil {
		return nil, err
	}
	
	// Calculate time cutoff for days filter
	cutoffTime := time.Now().AddDate(0, 0, -days).Unix()
	
	var logs []*core.Log
	found := 0
	for _, doc := range result {
		if count > 0 && found >= count {
			break
		}
		
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && 
		   log.ColonyName == colonyName && 
		   log.Timestamp >= cutoffTime &&
		   strings.Contains(strings.ToLower(log.Message), strings.ToLower(text)) {
			logs = append(logs, log)
			found++
		}
	}
	
	return logs, nil
}

func (a *ColonyOSAdapter) RemoveLogByID(logID string) error {
	return a.db.Delete(context.Background(), LogsCollection, logID)
}

func (a *ColonyOSAdapter) RemoveLogsByColonyName(colonyName string) error {
	result, err := a.db.List(context.Background(), LogsCollection, 10000, 0)
	if err != nil {
		return err
	}
	
	for _, doc := range result {
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ColonyName == colonyName {
			if err := a.db.Delete(context.Background(), LogsCollection, doc.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveLogsByProcessID(processID string) error {
	result, err := a.db.List(context.Background(), LogsCollection, 10000, 0)
	if err != nil {
		return err
	}
	
	for _, doc := range result {
		log, err := a.fieldsToLog(doc.Fields)
		if err == nil && log.ProcessID == processID {
			if err := a.db.Delete(context.Background(), LogsCollection, doc.ID); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (a *ColonyOSAdapter) RemoveAllLogs() error {
	result, err := a.db.List(context.Background(), LogsCollection, 10000, 0)
	if err != nil {
		return err
	}
	
	for _, doc := range result {
		if err := a.db.Delete(context.Background(), LogsCollection, doc.ID); err != nil {
			return err
		}
	}
	
	return nil
}

// Conversion helper methods

func (a *ColonyOSAdapter) logToFields(log *core.Log) map[string]interface{} {
	return map[string]interface{}{
		"process_id":    log.ProcessID,
		"colony_name":   log.ColonyName,
		"executor_name": log.ExecutorName,
		"timestamp":     log.Timestamp,
		"message":       log.Message,
	}
}

func (a *ColonyOSAdapter) fieldsToLog(fields map[string]interface{}) (*core.Log, error) {
	log := &core.Log{}
	
	if processID, ok := fields["process_id"].(string); ok {
		log.ProcessID = processID
	}
	if colonyName, ok := fields["colony_name"].(string); ok {
		log.ColonyName = colonyName
	}
	if executorName, ok := fields["executor_name"].(string); ok {
		log.ExecutorName = executorName
	}
	if timestamp, ok := fields["timestamp"].(int64); ok {
		log.Timestamp = timestamp
	} else if timestamp, ok := fields["timestamp"].(float64); ok {
		log.Timestamp = int64(timestamp)
	}
	if message, ok := fields["message"].(string); ok {
		log.Message = message
	}
	
	return log, nil
}