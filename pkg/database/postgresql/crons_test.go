package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestCronClosedDB(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	db.Close()

	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()

	err = db.AddCron(cron)
	assert.NotNil(t, err)

	err = db.UpdateCron("invalid_id", time.Now(), time.Time{}, core.GenerateRandomID())
	assert.NotNil(t, err)

	_, err = db.GetCronByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.FindCronsByColonyName("invalid_colony_name", 1)
	assert.NotNil(t, err)

	_, err = db.FindAllCrons()
	assert.NotNil(t, err)

	err = db.RemoveCronByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveAllCronsByColonyName("invalid_colony_name")
	assert.NotNil(t, err)
}

func TestAddCron(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()

	err = db.AddCron(cron)
	assert.Nil(t, err)

	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.NotNil(t, cronFromDB)
	assert.True(t, cron.Equals(cronFromDB))
}

func TestUpdateCron(t *testing.T) {
	db, err := PrepareTests()
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

	err = db.UpdateCron(cron.ID, time.Now(), time.Time{}, core.GenerateRandomID())
	assert.Nil(t, err)

	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Greater(t, cronFromDB.NextRun.Unix(), time.Time{}.Unix())
	assert.Equal(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
	assert.NotEqual(t, cronFromDB.PrevProcessGraphID, "")

	err = db.UpdateCron(cron.ID, time.Now(), time.Now(), core.GenerateRandomID())
	assert.Nil(t, err)
	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Greater(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
}

func TestFindCronsByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	cron1 := core.CreateCron(colonyName1, "test_name1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyName2, "test_name2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyName2, "test_name3", "* * * * * *", 0, false, "workflow3")
	cron3.ID = core.GenerateRandomID()

	err = db.AddCron(cron1)
	assert.Nil(t, err)
	err = db.AddCron(cron2)
	assert.Nil(t, err)
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	crons, err := db.FindCronsByColonyName(colonyName1, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 1)
	assert.Equal(t, crons[0].ID, cron1.ID)

	crons, err = db.FindCronsByColonyName(colonyName2, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 2)

	crons, err = db.FindCronsByColonyName(colonyName2, 1)
	assert.Len(t, crons, 1)
}

func TestFindAllCrons(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	cron1 := core.CreateCron(colonyName1, "test_name1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyName2, "test_name2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyName2, "test_name3", "* * * * * *", 0, false, "workflow3")
	cron3.ID = core.GenerateRandomID()

	err = db.AddCron(cron1)
	assert.Nil(t, err)
	err = db.AddCron(cron2)
	assert.Nil(t, err)
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	crons, err := db.FindAllCrons()
	assert.Nil(t, err)
	assert.Len(t, crons, 3)
}

func TestRemoveCronByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	cron := core.CreateCron(core.GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.ID = core.GenerateRandomID()
	err = db.AddCron(cron)
	assert.Nil(t, err)

	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Equal(t, cronFromDB.ID, cron.ID)

	err = db.RemoveCronByID(cron.ID)
	assert.Nil(t, err)

	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Nil(t, cronFromDB)
}

func TestRemoveAllCronsByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	cron1 := core.CreateCron(colonyName1, "test_name1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyName2, "test_name2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyName2, "test_name3", "* * * * * *", 0, false, "workflow3")
	cron3.ID = core.GenerateRandomID()

	err = db.AddCron(cron1)
	assert.Nil(t, err)
	err = db.AddCron(cron2)
	assert.Nil(t, err)
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	err = db.RemoveAllCronsByColonyName(colonyName2)
	assert.Nil(t, err)

	crons, err := db.FindCronsByColonyName(colonyName1, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 1)
	assert.Equal(t, crons[0].ID, cron1.ID)

	crons, err = db.FindCronsByColonyName(colonyName2, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 0)
}

func TestAddDuplicateCronRejected(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()

	// Create first cron
	cron1 := core.CreateCron(colonyName, "duplicate_name", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	// Attempt to create second cron with same name in same colony - should fail
	cron2 := core.CreateCron(colonyName, "duplicate_name", "0 * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify only one cron exists
	crons, err := db.FindCronsByColonyName(colonyName, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 1)
	assert.Equal(t, crons[0].ID, cron1.ID)
}

func TestSameCronNameDifferentColoniesAllowed(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()
	sharedName := "shared_cron_name"

	// Create cron in first colony
	cron1 := core.CreateCron(colonyName1, sharedName, "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	err = db.AddCron(cron1)
	assert.Nil(t, err)

	// Create cron with same name in second colony - should succeed
	cron2 := core.CreateCron(colonyName2, sharedName, "0 * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	err = db.AddCron(cron2)
	assert.Nil(t, err)

	// Verify both crons exist
	crons1, err := db.FindCronsByColonyName(colonyName1, 100)
	assert.Nil(t, err)
	assert.Len(t, crons1, 1)
	assert.Equal(t, crons1[0].Name, sharedName)

	crons2, err := db.FindCronsByColonyName(colonyName2, 100)
	assert.Nil(t, err)
	assert.Len(t, crons2, 1)
	assert.Equal(t, crons2[0].Name, sharedName)
}

func TestAddCronConcurrentDuplicateRejected(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	cronName := "concurrent_cron"

	// Launch multiple goroutines trying to create the same cron simultaneously
	numGoroutines := 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			cron := core.CreateCron(colonyName, cronName, "* * * * * *", 0, false, "workflow")
			cron.ID = core.GenerateRandomID()
			results <- db.AddCron(cron)
		}()
	}

	// Collect results
	successCount := 0
	failCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		} else {
			failCount++
			assert.Contains(t, err.Error(), "already exists")
		}
	}

	// Exactly one should succeed, the rest should fail due to unique constraint
	assert.Equal(t, 1, successCount, "Expected exactly one successful insert")
	assert.Equal(t, numGoroutines-1, failCount, "Expected all other inserts to fail")

	// Verify only one cron exists
	crons, err := db.FindCronsByColonyName(colonyName, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 1)
}

func TestGetCronByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	// Create crons with different names in different colonies
	cron1 := core.CreateCron(colonyName1, "reconcile-Database-dc1", "* * * * * *", 60, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyName1, "reconcile-Database-dc2", "* * * * * *", 60, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyName2, "reconcile-Database-dc1", "* * * * * *", 60, false, "workflow3")
	cron3.ID = core.GenerateRandomID()

	err = db.AddCron(cron1)
	assert.Nil(t, err)
	err = db.AddCron(cron2)
	assert.Nil(t, err)
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	// Test: Find cron by name in colony1
	foundCron, err := db.GetCronByName(colonyName1, "reconcile-Database-dc1")
	assert.Nil(t, err)
	assert.NotNil(t, foundCron)
	assert.Equal(t, cron1.ID, foundCron.ID)
	assert.Equal(t, cron1.Name, foundCron.Name)
	assert.Equal(t, cron1.ColonyName, foundCron.ColonyName)

	// Test: Find different cron by name in colony1
	foundCron, err = db.GetCronByName(colonyName1, "reconcile-Database-dc2")
	assert.Nil(t, err)
	assert.NotNil(t, foundCron)
	assert.Equal(t, cron2.ID, foundCron.ID)

	// Test: Same name in different colony returns different cron
	foundCron, err = db.GetCronByName(colonyName2, "reconcile-Database-dc1")
	assert.Nil(t, err)
	assert.NotNil(t, foundCron)
	assert.Equal(t, cron3.ID, foundCron.ID)
	assert.Equal(t, colonyName2, foundCron.ColonyName)

	// Test: Non-existent cron name returns nil
	foundCron, err = db.GetCronByName(colonyName1, "nonexistent-cron")
	assert.Nil(t, err)
	assert.Nil(t, foundCron)

	// Test: Non-existent colony returns nil
	foundCron, err = db.GetCronByName("nonexistent-colony", "reconcile-Database-dc1")
	assert.Nil(t, err)
	assert.Nil(t, foundCron)
}
