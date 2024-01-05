package postgresql

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddGetLogsByProcessID(t *testing.T) {
	db, err := PrepareTests()
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
}

func TestAddGetLogsByExecutor(t *testing.T) {
	db, err := PrepareTests()
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
}

func TestAddGetLogsByProcessIDSince(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	timestamp1 := time.Now().UTC().UnixNano()
	time.Sleep(2 * time.Second)
	timestamp2 := time.Now().UTC().UnixNano()

	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp1, "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp2, "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessIDSince("test_processid", 100, timestamp1)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestAddGetLogsByExecutorSince(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	timestamp1 := time.Now().UTC().UnixNano()
	time.Sleep(2 * time.Second)
	timestamp2 := time.Now().UTC().UnixNano()

	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp1, "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executor_name", timestamp2, "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByExecutorSince("test_executor_name", 100, timestamp1)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestRemoveLogsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	err = db.AddLog("test_processid1", "test_colony1", "test_executor_name", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid2", "test_colony2", "test_executor_name", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)

	err = db.RemoveLogsByColonyName("test_colony1")
	assert.Nil(t, err)

	logs, err = db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.GetLogsByProcessID("test_processid2", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestSearchLogs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	timestamp := time.Now().UTC()
	err = db.addHistoricalLog("test_processid1", "test_colony", "test_executor_name1", timestamp.UnixNano(), "1", timestamp)
	assert.Nil(t, err)
	timestamp = time.Now().UTC()
	err = db.addHistoricalLog("test_processid2", "test_colony", "test_executor_name1", timestamp.UnixNano(), "test", timestamp)
	assert.Nil(t, err)
	timestamp = time.Now().UTC()
	err = db.addHistoricalLog("test_processid2", "test_colony", "test_executor_name1", timestamp.UnixNano(), "test", timestamp)
	assert.Nil(t, err)
	timestamp = time.Now().Add(-24 * time.Hour * 3).UTC()
	err = db.addHistoricalLog("test_processid3", "test_colony", "test_executor_name2", timestamp.UnixNano(), "error", timestamp)
	assert.Nil(t, err)

	logs, err := db.SearchLogs("test_colony", "", -1, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.SearchLogs("test_colony", "error", -1, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.SearchLogs("test_colony", "error", 0, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.SearchLogs("test_colony", "error", 1, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.SearchLogs("test_colony", "error", 2, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.SearchLogs("test_colony", "error", 3, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.SearchLogs("test_colony", "error", 4, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	for _, log := range logs {
		assert.Equal(t, log.ProcessID, "test_processid3")
		assert.Equal(t, log.Message, "error")
	}

	logs, err = db.SearchLogs("test_colony", "test", 3, 10)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)
	for _, log := range logs {
		assert.Equal(t, log.Message, "test")
		assert.Equal(t, log.ExecutorName, "test_executor_name1")
	}
}
