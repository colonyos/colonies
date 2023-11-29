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

	err = db.AddLog("test_processid", "test_colony", "test_executorid", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executorid", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, logs[0].ProcessID, "test_processid")
	assert.Equal(t, logs[0].ColonyName, "test_colony")
	assert.Equal(t, logs[0].ExecutorID, "test_executorid")
}

func TestAddGetLogsByExecutorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	err = db.AddLog("test_processid", "test_colony", "test_executorid1", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executorid2", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByExecutorID("test_executorid2", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
	assert.Equal(t, logs[0].ExecutorID, "test_executorid2")
}

func TestAddGetLogsByProcessIDSince(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	timestamp1 := time.Now().UTC().UnixNano()
	time.Sleep(2 * time.Second)
	timestamp2 := time.Now().UTC().UnixNano()

	err = db.AddLog("test_processid", "test_colony", "test_executorid", timestamp1, "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executorid", timestamp2, "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessIDSince("test_processid", 100, timestamp1)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestAddGetLogsByExecutorIDSince(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	timestamp1 := time.Now().UTC().UnixNano()
	time.Sleep(2 * time.Second)
	timestamp2 := time.Now().UTC().UnixNano()

	err = db.AddLog("test_processid", "test_colony", "test_executorid", timestamp1, "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colony", "test_executorid", timestamp2, "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByExecutorIDSince("test_executorid", 100, timestamp1)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestDeleteLogsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	err = db.AddLog("test_processid1", "test_colony1", "test_executorid", time.Now().UTC().UnixNano(), "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid2", "test_colony2", "test_executorid", time.Now().UTC().UnixNano(), "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)

	err = db.DeleteLogsByColonyName("test_colony1")
	assert.Nil(t, err)

	logs, err = db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.GetLogsByProcessID("test_processid2", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}
