package kvstore

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestCronClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()

	// KVStore operations work even after close (in-memory store)
	err = db.AddCron(cron)
	assert.Nil(t, err)

	err = db.UpdateCron("invalid_id", time.Now(), time.Time{}, core.GenerateRandomID())
	assert.NotNil(t, err) // Should error for non-existing cron

	// The cron we added should be retrievable
	_, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)

	_, err = db.FindCronsByColonyName("invalid_colony_name", 1)
	assert.Nil(t, err) // Returns empty slice

	_, err = db.FindAllCrons()
	assert.Nil(t, err)

	err = db.RemoveCronByID("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing cron

	err = db.RemoveAllCronsByColonyName("invalid_colony_name")
	assert.Nil(t, err) // No error when nothing to remove
}

func TestAddCron(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test adding nil cron
	err = db.AddCron(nil)
	assert.NotNil(t, err)

	// Create and add valid cron
	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()

	err = db.AddCron(cron)
	assert.Nil(t, err)

	// Test duplicate cron
	err = db.AddCron(cron)
	assert.NotNil(t, err)

	// Verify cron was added
	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB)
	assert.True(t, cron.Equals(cronFromDB))
}

func TestGetCronByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add cron
	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()
	err = db.AddCron(cron)
	assert.Nil(t, err)

	// Get existing cron
	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB)
	assert.True(t, cron.Equals(cronFromDB))

	// Test non-existing cron
	nonExistingCron, err := db.GetCronByID("non_existing_id")
	assert.Nil(t, err)
	assert.Nil(t, nonExistingCron)
}

func TestGetCronByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	cron := core.CreateCron(colonyName, "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()
	err = db.AddCron(cron)
	assert.Nil(t, err)

	// Get by name
	cronFromDB, err := db.GetCronByName(colonyName, "test_name")
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB)
	assert.True(t, cron.Equals(cronFromDB))

	// Test non-existing cron - should return (nil, nil) like PostgreSQL
	nonExistingCron, err := db.GetCronByName(colonyName, "non_existing_name")
	assert.Nil(t, err)
	assert.Nil(t, nonExistingCron)

	// Test invalid colony - should return (nil, nil) like PostgreSQL
	invalidCron, err := db.GetCronByName("invalid_colony", "test_name")
	assert.Nil(t, err)
	assert.Nil(t, invalidCron)
}

func TestUpdateCron(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	cron := core.CreateCron(colonyName, "test_name", "* * * * * *", 100, true, "workflow")
	cron.ID = core.GenerateRandomID()

	err = db.AddCron(cron)
	assert.Nil(t, err)

	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Equal(t, cronFromDB.ID, cron.ID)
	assert.Equal(t, cronFromDB.ColonyName, colonyName)
	assert.Equal(t, cronFromDB.Name, "test_name")
	assert.Equal(t, cronFromDB.CronExpression, "* * * * * *")
	assert.Equal(t, cronFromDB.Interval, 100)
	assert.Equal(t, cronFromDB.Random, true)
	assert.Equal(t, cronFromDB.WorkflowSpec, "workflow")
	assert.Equal(t, cronFromDB.PrevProcessGraphID, "")

	// Update cron
	nextRun := time.Now().Add(time.Hour)
	processGraphID := core.GenerateRandomID()
	err = db.UpdateCron(cron.ID, nextRun, time.Time{}, processGraphID)
	assert.Nil(t, err)

	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Equal(t, cronFromDB.NextRun.Unix(), nextRun.Unix())
	assert.Equal(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
	assert.Equal(t, cronFromDB.PrevProcessGraphID, processGraphID)

	// Update with last run
	lastRun := time.Now()
	err = db.UpdateCron(cron.ID, nextRun, lastRun, processGraphID)
	assert.Nil(t, err)
	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Equal(t, cronFromDB.LastRun.Unix(), lastRun.Unix())

	// Test updating non-existing cron
	err = db.UpdateCron("invalid_id", time.Now(), time.Time{}, core.GenerateRandomID())
	assert.NotNil(t, err)
}

func TestFindCronsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony1 := core.GenerateRandomID()
	colony2 := core.GenerateRandomID()

	// Add crons to different colonies
	cron1 := core.CreateCron(colony1, "cron1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	cron2 := core.CreateCron(colony1, "cron2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.Nil(t, err)

	cron3 := core.CreateCron(colony2, "cron3", "* * * * * *", 0, false, "workflow3")
	cron3.ID = core.GenerateRandomID()
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	// Find crons by colony
	crons1, err := db.FindCronsByColonyName(colony1, 10)
	assert.Nil(t, err)
	assert.Len(t, crons1, 2)

	crons2, err := db.FindCronsByColonyName(colony2, 10)
	assert.Nil(t, err)
	assert.Len(t, crons2, 1)

	// Test with count limit
	cronsLimited, err := db.FindCronsByColonyName(colony1, 1)
	assert.Nil(t, err)
	assert.Len(t, cronsLimited, 1)

	// Test non-existing colony
	cronsInvalid, err := db.FindCronsByColonyName("invalid_colony", 10)
	assert.Nil(t, err)
	assert.Empty(t, cronsInvalid)
}

func TestFindAllCrons(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add multiple crons
	cron1 := core.CreateCron("colony1", "cron1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	cron2 := core.CreateCron("colony2", "cron2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.Nil(t, err)

	// Find all crons
	allCrons, err := db.FindAllCrons()
	assert.Nil(t, err)
	assert.Len(t, allCrons, 2)

	// Verify crons are included
	cronIDs := make([]string, len(allCrons))
	for i, cron := range allCrons {
		cronIDs[i] = cron.ID
	}
	assert.Contains(t, cronIDs, cron1.ID)
	assert.Contains(t, cronIDs, cron2.ID)
}

func TestRemoveCron(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Add cron
	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()
	err = db.AddCron(cron)
	assert.Nil(t, err)

	// Verify cron exists
	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB)

	// Remove cron
	err = db.RemoveCronByID(cron.ID)
	assert.Nil(t, err)

	// Verify cron is gone
	removedCron, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Nil(t, removedCron)

	// Test removing non-existing cron
	err = db.RemoveCronByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestRemoveAllCronsByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.GenerateRandomID()

	// Add multiple crons to the colony
	cron1 := core.CreateCron(colony, "cron1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	cron2 := core.CreateCron(colony, "cron2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.Nil(t, err)

	// Add cron to different colony
	otherCron := core.CreateCron("other_colony", "other_cron", "* * * * * *", 0, false, "workflow3")
	otherCron.ID = core.GenerateRandomID()
	err = db.AddCron(otherCron)
	assert.Nil(t, err)

	// Verify crons exist
	crons, err := db.FindCronsByColonyName(colony, 10)
	assert.Nil(t, err)
	assert.Len(t, crons, 2)

	// Remove all crons from specific colony
	err = db.RemoveAllCronsByColonyName(colony)
	assert.Nil(t, err)

	// Verify crons from specific colony are gone
	remainingCrons, err := db.FindCronsByColonyName(colony, 10)
	assert.Nil(t, err)
	assert.Empty(t, remainingCrons)

	// Verify other colony's cron still exists
	otherColonyCrons, err := db.FindCronsByColonyName("other_colony", 10)
	assert.Nil(t, err)
	assert.Len(t, otherColonyCrons, 1)

	// Test removing from invalid colony - should not error
	err = db.RemoveAllCronsByColonyName("invalid_colony")
	assert.Nil(t, err)
}

func TestCronComplexWorkflow(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colony := core.GenerateRandomID()

	// Create cron with complex workflow specification
	workflowSpec := `{
		"name": "test_workflow",
		"steps": [
			{"name": "step1", "command": "echo hello"},
			{"name": "step2", "command": "echo world", "depends_on": ["step1"]}
		]
	}`

	cron := core.CreateCron(colony, "complex_cron", "0 0 * * *", 3600, true, workflowSpec)
	cron.ID = core.GenerateRandomID()
	
	// Set additional properties
	nextRun := time.Now().Add(24 * time.Hour)
	cron.NextRun = nextRun

	err = db.AddCron(cron)
	assert.Nil(t, err)

	// Retrieve and verify
	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB)
	assert.Equal(t, cronFromDB.WorkflowSpec, workflowSpec)
	assert.Equal(t, cronFromDB.CronExpression, "0 0 * * *")
	assert.Equal(t, cronFromDB.Interval, 3600)
	assert.True(t, cronFromDB.Random)
	assert.Equal(t, cronFromDB.NextRun.Unix(), nextRun.Unix())
}