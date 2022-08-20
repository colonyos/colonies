package core

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateCron(t *testing.T) {
	cron := CreateCron(GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	assert.Equal(t, cron.Name, "test_name")
	assert.Equal(t, cron.CronExpression, "* * * * * *")
	assert.Equal(t, cron.WorkflowSpec, "workflow")
}

func TestIsCronEquals(t *testing.T) {
	cron1 := CreateCron(GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	cron2 := CreateCron(GenerateRandomID(), "test_name2", "* * * * * *", 0, false, "workflow2")
	cron3 := CreateCron(GenerateRandomID(), "test_name3", "* * * * * *", 0, false, "workflow3")

	assert.True(t, cron1.Equals(cron1))
	assert.False(t, cron1.Equals(cron2))
	assert.False(t, cron1.Equals(cron3))
}

func TestIsCronArraysEquals(t *testing.T) {
	cron1 := CreateCron(GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	cron2 := CreateCron(GenerateRandomID(), "test_name2", "* * * * * *", 0, false, "workflow2")
	cron3 := CreateCron(GenerateRandomID(), "test_name3", "* * * * * *", 0, false, "workflow3")
	cron4 := CreateCron(GenerateRandomID(), "test_name4", "* * * * * *", 0, false, "workflow4")
	cron5 := CreateCron(GenerateRandomID(), "test_name5", "* * * * * *", 0, false, "workflow5")

	var crons1 []*Cron
	var crons2 []*Cron
	var crons3 []*Cron

	crons1 = append(crons1, cron1)
	crons1 = append(crons1, cron2)
	crons1 = append(crons1, cron3)

	crons2 = append(crons2, cron2)
	crons2 = append(crons2, cron1)
	crons2 = append(crons2, cron3)

	crons3 = append(crons3, cron1)
	crons3 = append(crons3, cron2)
	crons3 = append(crons3, cron3)
	crons3 = append(crons3, cron4)
	crons3 = append(crons3, cron5)

	assert.True(t, IsCronArraysEqual(crons1, crons1))
	assert.True(t, IsCronArraysEqual(crons1, crons2))
	assert.False(t, IsCronArraysEqual(crons1, crons3))
}

func TestCronToJSON(t *testing.T) {
	cron := CreateCron(GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	jsonStr, err := cron.ToJSON()
	assert.Nil(t, err)

	fmt.Println(jsonStr)

	cron2, err := ConvertJSONToCron(jsonStr)
	assert.Nil(t, err)
	assert.True(t, cron.Equals(cron2))
}

func TestCronArrayToJSON(t *testing.T) {
	cron1 := CreateCron(GenerateRandomID(), "test_name1", "* * * * * *", 0, false, "workflow1")
	cron2 := CreateCron(GenerateRandomID(), "test_name2", "* * * * * *", 0, false, "workflow2")
	cron3 := CreateCron(GenerateRandomID(), "test_name3", "* * * * * *", 0, false, "workflow3")

	var crons []*Cron
	crons = append(crons, cron1)
	crons = append(crons, cron2)
	crons = append(crons, cron3)

	jsonStr, err := ConvertCronArrayToJSON(crons)
	assert.Nil(t, err)

	crons2, err := ConvertJSONToCronArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsCronArraysEqual(crons, crons2))
}

func TestCronHasExpire(t *testing.T) {
	cron := CreateCron(GenerateRandomID(), "test_name", "* * * * * *", 0, false, "workflow")
	cron.NextRun = time.Now().Add(-100 * time.Second)
	assert.True(t, cron.HasExpired())
	cron.NextRun = time.Now().Add(100 * time.Second)
	assert.False(t, cron.HasExpired())
}
