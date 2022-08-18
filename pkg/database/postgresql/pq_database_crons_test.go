package postgresql

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

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

	colonyID := core.GenerateRandomID()
	cron := core.CreateCron(colonyID, "test_name", "* * * * * *", 100, true, "workflow")
	cron.ID = core.GenerateRandomID()

	err = db.AddCron(cron)
	assert.Nil(t, err)

	cronFromDB, err := db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Equal(t, cronFromDB.ID, cron.ID)
	assert.Equal(t, cronFromDB.ColonyID, colonyID)
	assert.Equal(t, cronFromDB.Name, "test_name")
	assert.Equal(t, cronFromDB.CronExpression, "* * * * * *")
	assert.Equal(t, cronFromDB.Intervall, 100)
	assert.Equal(t, cronFromDB.Random, true)
	assert.Equal(t, cronFromDB.WorkflowSpec, "workflow")
	assert.Equal(t, cronFromDB.LastProcessGraphID, "")

	err = db.UpdateCron(cron.ID, time.Now(), time.Time{}, core.GenerateRandomID())
	assert.Nil(t, err)

	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Greater(t, cronFromDB.NextRun.Unix(), time.Time{}.Unix())
	assert.Equal(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
	assert.NotEqual(t, cronFromDB.LastProcessGraphID, "")

	err = db.UpdateCron(cron.ID, time.Now(), time.Now(), core.GenerateRandomID())
	assert.Nil(t, err)
	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Greater(t, cronFromDB.LastRun.Unix(), time.Time{}.Unix())
}

func TestFindCronsByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID1 := core.GenerateRandomID()
	colonyID2 := core.GenerateRandomID()

	cron1 := core.CreateCron(colonyID1, "test_name1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyID2, "test_name2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyID2, "test_name3", "* * * * * *", 0, false, "workflow3")
	cron3.ID = core.GenerateRandomID()

	err = db.AddCron(cron1)
	assert.Nil(t, err)
	err = db.AddCron(cron2)
	assert.Nil(t, err)
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	crons, err := db.FindCronsByColonyID(colonyID1, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 1)
	assert.Equal(t, crons[0].ID, cron1.ID)

	crons, err = db.FindCronsByColonyID(colonyID2, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 2)

	crons, err = db.FindCronsByColonyID(colonyID2, 1)
	assert.Len(t, crons, 1)
}

func TestFindAllCrons(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID1 := core.GenerateRandomID()
	colonyID2 := core.GenerateRandomID()

	cron1 := core.CreateCron(colonyID1, "test_name1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyID2, "test_name2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyID2, "test_name3", "* * * * * *", 0, false, "workflow3")
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

func TestDeleteCronByID(t *testing.T) {
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

	err = db.DeleteCronByID(cron.ID)
	assert.Nil(t, err)

	cronFromDB, err = db.GetCronByID(cron.ID)
	assert.Nil(t, err)
	assert.Nil(t, cronFromDB)
}

func TestDeleteAllCronsByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID1 := core.GenerateRandomID()
	colonyID2 := core.GenerateRandomID()

	cron1 := core.CreateCron(colonyID1, "test_name1", "* * * * * *", 0, false, "workflow1")
	cron1.ID = core.GenerateRandomID()
	cron2 := core.CreateCron(colonyID2, "test_name2", "* * * * * *", 0, false, "workflow2")
	cron2.ID = core.GenerateRandomID()
	cron3 := core.CreateCron(colonyID2, "test_name3", "* * * * * *", 0, false, "workflow3")
	cron3.ID = core.GenerateRandomID()

	err = db.AddCron(cron1)
	assert.Nil(t, err)
	err = db.AddCron(cron2)
	assert.Nil(t, err)
	err = db.AddCron(cron3)
	assert.Nil(t, err)

	err = db.DeleteAllCronsByColonyID(colonyID2)
	assert.Nil(t, err)

	crons, err := db.FindCronsByColonyID(colonyID1, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 1)
	assert.Equal(t, crons[0].ID, cron1.ID)

	crons, err = db.FindCronsByColonyID(colonyID2, 100)
	assert.Nil(t, err)
	assert.Len(t, crons, 0)
}
