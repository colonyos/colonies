package kvstore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	// KVStore operations work even after close (in-memory store)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name", time.Now().UTC().UnixNano(), "test message")
	assert.Nil(t, err)

	_, err = db.GetLogsByProcessID("test_processid", 100)
	assert.Nil(t, err)

	_, err = db.GetLogsByProcessIDSince("test_processid", 100, time.Now().UnixNano())
	assert.Nil(t, err)

	_, err = db.GetLogsByExecutor("test_executor", 100)
	assert.Nil(t, err)

	_, err = db.GetLogsByExecutorSince("test_executor", 100, time.Now().UnixNano())
	assert.Nil(t, err)

	err = db.RemoveLogsByColonyName("test_colony")
	assert.Nil(t, err)

	_, err = db.CountLogs("test_colony")
	assert.Nil(t, err)

	_, err = db.SearchLogs("test_colony", "test", 7, 100)
	assert.Nil(t, err)
}

func TestAddGetLogsByProcessID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	err = db.AddLog("test_processid", "test_colony", "test_executor_name", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, logs[0].ProcessID, "test_processid")
	assert.Equal(t, logs[0].ColonyName, "test_colony")
	assert.Equal(t, logs[0].ExecutorName, "test_executor_name")

	// Test with non-existing process ID
	emptyLogs, err := db.GetLogsByProcessID("non_existing_id", 100)
	assert.Nil(t, err)
	assert.Empty(t, emptyLogs)

	// Test with limit
	limitedLogs, err := db.GetLogsByProcessID("test_processid", 1)
	assert.Nil(t, err)
	assert.Len(t, limitedLogs, 1)
}

func TestAddGetLogsByExecutor(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	err = db.AddLog("test_processid", "test_colony", "test_executor_name1", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name2", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByExecutor("test_executor_name2", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].ExecutorName, "test_executor_name2")

	// Test with non-existing executor
	emptyLogs, err := db.GetLogsByExecutor("non_existing_executor", 100)
	assert.Nil(t, err)
	assert.Empty(t, emptyLogs)

	// Test all logs for executor1
	logs1, err := db.GetLogsByExecutor("test_executor_name1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs1, 1)
	assert.Equal(t, logs1[0].Message, "1")
}

func TestAddGetLogsByProcessIDSince(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	timestamp1 := time.Now().UTC().UnixNano()
	time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamps
	timestamp2 := time.Now().UTC().UnixNano()

	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp1, "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp2, "2")
	assert.Nil(t, err)

	// Get logs since timestamp1 (should get only the second log)
	logs, err := db.GetLogsByProcessIDSince("test_processid", 100, timestamp1)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "2")

	// Get logs since timestamp2 (should get no logs)
	logsEmpty, err := db.GetLogsByProcessIDSince("test_processid", 100, timestamp2)
	assert.Nil(t, err)
	assert.Empty(t, logsEmpty)

	// Get logs since a timestamp before timestamp1 (should get both logs)
	logsBoth, err := db.GetLogsByProcessIDSince("test_processid", 100, timestamp1-1000000)
	assert.Nil(t, err)
	assert.Len(t, logsBoth, 2)
}

func TestAddGetLogsByExecutorSince(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	timestamp1 := time.Now().UTC().UnixNano()
	time.Sleep(10 * time.Millisecond)
	timestamp2 := time.Now().UTC().UnixNano()

	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp1, "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp2, "2")
	assert.Nil(t, err)

	// Get logs since timestamp1 for executor
	logs, err := db.GetLogsByExecutorSince("test_executor_name", 100, timestamp1)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, "2")

	// Test with non-existing executor
	emptyLogs, err := db.GetLogsByExecutorSince("non_existing_executor", 100, timestamp1)
	assert.Nil(t, err)
	assert.Empty(t, emptyLogs)
}

func TestRemoveLogsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	err = db.AddLog("test_processid1", "test_colony1", "test_executor_name", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid2", "test_colony2", "test_executor_name", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)
	err = db.AddLog("test_processid3", "test_colony1", "test_executor_name", time.Now().UTC().UnixNano(), "3")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)

	logs2, err := db.GetLogsByProcessID("test_processid2", 100)
	assert.Nil(t, err)
	assert.Len(t, logs2, 1)

	// Remove logs for colony1
	err = db.RemoveLogsByColonyName("test_colony1")
	assert.Nil(t, err)

	// Verify colony1 logs are gone
	logsAfterRemoval1, err := db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Empty(t, logsAfterRemoval1)

	logsAfterRemoval3, err := db.GetLogsByProcessID("test_processid3", 100)
	assert.Nil(t, err)
	assert.Empty(t, logsAfterRemoval3)

	// Verify colony2 logs still exist
	logsAfterRemoval2, err := db.GetLogsByProcessID("test_processid2", 100)
	assert.Nil(t, err)
	assert.Len(t, logsAfterRemoval2, 1)

	// Test removing from non-existing colony - should not error
	err = db.RemoveLogsByColonyName("non_existing_colony")
	assert.Nil(t, err)
}

func TestCountLogs(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add logs to different colonies
	err = db.AddLog("process1", "colony1", "executor1", time.Now().UTC().UnixNano(), "message1")
	assert.Nil(t, err)
	err = db.AddLog("process2", "colony1", "executor2", time.Now().UTC().UnixNano(), "message2")
	assert.Nil(t, err)
	err = db.AddLog("process3", "colony2", "executor3", time.Now().UTC().UnixNano(), "message3")
	assert.Nil(t, err)

	// Count logs by colony
	count1, err := db.CountLogs("colony1")
	assert.Nil(t, err)
	assert.Equal(t, count1, 2)

	count2, err := db.CountLogs("colony2")
	assert.Nil(t, err)
	assert.Equal(t, count2, 1)

	// Count logs for non-existing colony
	countEmpty, err := db.CountLogs("non_existing_colony")
	assert.Nil(t, err)
	assert.Equal(t, countEmpty, 0)
}

func TestSearchLogs(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add logs with different messages
	err = db.AddLog("process1", "test_colony", "executor1", time.Now().UTC().UnixNano(), "error: failed to connect")
	assert.Nil(t, err)
	err = db.AddLog("process2", "test_colony", "executor2", time.Now().UTC().UnixNano(), "info: connection established")
	assert.Nil(t, err)
	err = db.AddLog("process3", "test_colony", "executor3", time.Now().UTC().UnixNano(), "warning: connection timeout")
	assert.Nil(t, err)
	err = db.AddLog("process4", "other_colony", "executor4", time.Now().UTC().UnixNano(), "error: failed to start")
	assert.Nil(t, err)

	// Search for logs containing "error" in test_colony
	errorLogs, err := db.SearchLogs("test_colony", "error", 7, 100)
	assert.Nil(t, err)
	assert.Len(t, errorLogs, 1)
	assert.Equal(t, errorLogs[0].ProcessID, "process1")

	// Search for logs containing "connection" in test_colony
	connectionLogs, err := db.SearchLogs("test_colony", "connection", 7, 100)
	assert.Nil(t, err)
	assert.Len(t, connectionLogs, 2)

	// Search with limit
	limitedLogs, err := db.SearchLogs("test_colony", "connection", 7, 1)
	assert.Nil(t, err)
	assert.Len(t, limitedLogs, 1)

	// Search for non-existing text
	noLogs, err := db.SearchLogs("test_colony", "nonexistent", 7, 100)
	assert.Nil(t, err)
	assert.Empty(t, noLogs)

	// Search in non-existing colony
	noColonyLogs, err := db.SearchLogs("non_existing_colony", "error", 7, 100)
	assert.Nil(t, err)
	assert.Empty(t, noColonyLogs)
}

func TestLogTimestamps(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test with specific timestamps
	timestamp1 := int64(1640995200000000000) // 2022-01-01 00:00:00 UTC
	timestamp2 := int64(1640995260000000000) // 2022-01-01 00:01:00 UTC

	err = db.AddLog("process1", "test_colony", "executor1", timestamp1, "first message")
	assert.Nil(t, err)
	err = db.AddLog("process1", "test_colony", "executor1", timestamp2, "second message")
	assert.Nil(t, err)

	// Verify logs are ordered by timestamp
	logs, err := db.GetLogsByProcessID("process1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)
	
	// Should be ordered by timestamp (oldest first or newest first depending on implementation)
	assert.Equal(t, logs[0].Timestamp, timestamp1)
	assert.Equal(t, logs[1].Timestamp, timestamp2)
}

func TestLogComplexMessages(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test with complex log message
	complexMessage := `{
		"level": "error",
		"message": "Failed to process request",
		"details": {
			"error_code": 500,
			"stack_trace": "line1\nline2\nline3"
		}
	}`

	err = db.AddLog("process1", "test_colony", "executor1", time.Now().UTC().UnixNano(), complexMessage)
	assert.Nil(t, err)

	// Retrieve and verify
	logs, err := db.GetLogsByProcessID("process1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].Message, complexMessage)

	// Test search with partial JSON content
	searchResults, err := db.SearchLogs("test_colony", "error_code", 7, 100)
	assert.Nil(t, err)
	assert.Len(t, searchResults, 1)
}