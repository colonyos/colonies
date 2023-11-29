package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRetentionClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	err = db.ApplyRetentionPolicy(1000)
	assert.NotNil(t, err)
}

func TestCalcTimestamp(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	period := 60 * 60 * 24
	now, timestamp := db.calcTimestamp(int64(period))

	assert.Equal(t, now.Unix()-timestamp.Unix(), int64(period))
}

func TestApplyRetentionPolicy(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	err = db.AddLog("test_processid", "test_colonyid", "test_executorid", time.Now().UTC().UnixNano(), "test_msg")
	assert.Nil(t, err)

	colonyName := core.GenerateRandomID()

	process := utils.CreateTestProcess(colonyName)
	err = db.AddProcess(process)
	assert.Nil(t, err)

	attribute := core.CreateAttribute(process.ID, colonyName, "", core.IN, "test_key2", "test_value2")
	err = db.AddAttribute(attribute)
	assert.Nil(t, err)

	graph, err := core.CreateProcessGraph(colonyName)
	assert.Nil(t, err)
	graph.AddRoot(process.ID)
	err = db.AddProcessGraph(graph)
	assert.Nil(t, err)

	err = db.SetProcessState(process.ID, core.SUCCESS)
	assert.Nil(t, err)
	err = db.SetProcessGraphState(graph.ID, core.SUCCESS)
	assert.Nil(t, err)

	count, err := db.CountSuccessfulProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	logs, err := db.GetLogsByProcessID("test_processid", 100)
	assert.Len(t, logs, 1)

	err = db.ApplyRetentionPolicy(1) // has no effect, it has not passed 1 second yet
	assert.Nil(t, err)

	count, err = db.CountSuccessfulProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	count, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, count, 1)

	_, err = db.GetAttributeByID(attribute.ID)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	err = db.ApplyRetentionPolicy(1)
	assert.Nil(t, err)

	count, err = db.CountSuccessfulProcessGraphs()
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	count, err = db.CountSuccessfulProcesses()
	assert.Nil(t, err)
	assert.Equal(t, count, 0)

	_, err = db.GetAttributeByID(attribute.ID)
	assert.NotNil(t, err)

	logs, err = db.GetLogsByProcessID("test_processid", 100)
	assert.Len(t, logs, 0)

	defer db.Close()
}
