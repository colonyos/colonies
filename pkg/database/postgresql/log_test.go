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

	logStr, err := db.GetLogsByProcessID("test_processid", 100)
	assert.Nil(t, err)
	assert.Equal(t, logStr, "12")
}

func TestDeleteLogs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	err = db.AddLog("test_processid1", "test_colonyid1", "test_executorid", "1")
	assert.Nil(t, err)
	err = db.AddLog("test_processid2", "test_colonyid2", "test_executorid", "2")
	assert.Nil(t, err)

	logStr, err := db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Equal(t, logStr, "1")

	err = db.DeleteLogs("test_colonyid1")
	assert.Nil(t, err)

	logStr, err = db.GetLogsByProcessID("test_processid1", 100)
	assert.Nil(t, err)
	assert.Equal(t, logStr, "")

	logStr, err = db.GetLogsByProcessID("test_processid2", 100)
	assert.Nil(t, err)
	assert.Equal(t, logStr, "2")
}
