package embedded

import (
	"sort"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

func (db *EmbeddedDatabase) AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error {
	logID := core.GenerateRandomID()
	logEntry := &core.Log{
		ProcessID:    processID,
		ColonyName:   colonyName,
		ExecutorName: executorName,
		Timestamp:    timestamp,
		Message:      msg,
	}

	if err := db.logs.Put(logID, logEntry); err != nil {
		return err
	}

	db.logsIdx.byProcess.Add(logID, processID)
	db.logsIdx.byExecutor.Add(logID, executorName)
	db.logsIdx.byColony.Add(logID, colonyName)

	return nil
}

func (db *EmbeddedDatabase) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) {
	ids := db.logsIdx.byProcess.Lookup(processID)
	logs := db.collectLogs(ids)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp < logs[j].Timestamp
	})
	if len(logs) > limit {
		logs = logs[:limit]
	}
	return logs, nil
}

func (db *EmbeddedDatabase) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) {
	ids := db.logsIdx.byProcess.Lookup(processID)
	var logs []*core.Log
	for _, id := range ids {
		if l, ok := db.logs.Get(id); ok {
			if l.Timestamp > since {
				logs = append(logs, copyLog(l))
			}
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp < logs[j].Timestamp
	})
	if len(logs) > limit {
		logs = logs[:limit]
	}
	return logs, nil
}

func (db *EmbeddedDatabase) GetLogsByProcessIDLatest(processID string, limit int) ([]*core.Log, error) {
	ids := db.logsIdx.byProcess.Lookup(processID)
	logs := db.collectLogs(ids)
	// Sort descending to get latest first
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp > logs[j].Timestamp
	})
	if len(logs) > limit {
		logs = logs[:limit]
	}
	// Reverse for chronological order
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
	return logs, nil
}

func (db *EmbeddedDatabase) GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error) {
	ids := db.logsIdx.byExecutor.Lookup(executorName)
	logs := db.collectLogs(ids)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp < logs[j].Timestamp
	})
	if len(logs) > limit {
		logs = logs[:limit]
	}
	return logs, nil
}

func (db *EmbeddedDatabase) GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error) {
	ids := db.logsIdx.byExecutor.Lookup(executorName)
	var logs []*core.Log
	for _, id := range ids {
		if l, ok := db.logs.Get(id); ok {
			if l.Timestamp > since {
				logs = append(logs, copyLog(l))
			}
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp < logs[j].Timestamp
	})
	if len(logs) > limit {
		logs = logs[:limit]
	}
	return logs, nil
}

func (db *EmbeddedDatabase) GetLogsByExecutorLatest(executorName string, limit int) ([]*core.Log, error) {
	ids := db.logsIdx.byExecutor.Lookup(executorName)
	logs := db.collectLogs(ids)
	// Sort descending to get latest first
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp > logs[j].Timestamp
	})
	if len(logs) > limit {
		logs = logs[:limit]
	}
	// Reverse for chronological order
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
	return logs, nil
}

func (db *EmbeddedDatabase) RemoveLogsByColonyName(colonyName string) error {
	ids := db.logsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if l, ok := db.logs.Get(id); ok {
			db.logsIdx.byProcess.Remove(id, l.ProcessID)
			db.logsIdx.byExecutor.Remove(id, l.ExecutorName)
			db.logsIdx.byColony.Remove(id, l.ColonyName)
			db.logs.Delete(id)
		}
	}
	return nil
}

func (db *EmbeddedDatabase) CountLogs(colonyName string) (int, error) {
	return db.logsIdx.byColony.Count(colonyName), nil
}

func (db *EmbeddedDatabase) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	var result []*core.Log

	ids := db.logsIdx.byColony.Lookup(colonyName)
	for _, id := range ids {
		if len(result) >= count {
			break
		}
		if l, ok := db.logs.Get(id); ok {
			if strings.Contains(l.Message, text) {
				// Check if the log was added within the days window.
				// Since we don't store Added time in core.Log, use Timestamp as a proxy.
				// Convert timestamp (which is a Unix timestamp in nanoseconds or seconds)
				// For compatibility, we accept the log if it's within the window.
				logTime := time.Unix(0, l.Timestamp)
				if logTime.IsZero() || logTime.Before(cutoff) {
					// Try treating as seconds
					logTime = time.Unix(l.Timestamp, 0)
				}
				if !logTime.Before(cutoff) {
					result = append(result, copyLog(l))
				}
			}
		}
	}

	return result, nil
}

func copyLog(l *core.Log) *core.Log {
	c := *l
	return &c
}

func (db *EmbeddedDatabase) collectLogs(ids []string) []*core.Log {
	var logs []*core.Log
	for _, id := range ids {
		if l, ok := db.logs.Get(id); ok {
			logs = append(logs, copyLog(l))
		}
	}
	return logs
}
