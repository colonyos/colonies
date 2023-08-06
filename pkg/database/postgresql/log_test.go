package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddGetLogs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	err = db.AddLog("test_processid", "test_colonyid", "test_executorid", "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid", "test_colonyid", "test_executorid", "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid", 100)
	assert.Len(t, logs, 2)
	assert.Equal(t, logs[0].ProcessID, "test_processid")
	assert.Equal(t, logs[0].ColonyID, "test_colonyid")
	assert.Equal(t, logs[0].ExecutorID, "test_executorid")
}

func TestDeleteLogs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	err = db.AddLog("test_processid1", "test_colonyid1", "test_executorid", "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid2", "test_colonyid2", "test_executorid", "2")
	assert.Nil(t, err)

	logs, err := db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)

	err = db.DeleteLogs("test_colonyid1")
	assert.Nil(t, err)

	logs, err = db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 0)

	logs, err = db.GetLogsByProcessID("test_processid2", 100)
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}
