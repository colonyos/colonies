package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCron(t *testing.T) {
	cron := CreateCron(GenerateRandomID(), "test_name", "* * * * * *", "workflow")
	assert.Len(t, cron.ID, 64)
	assert.Equal(t, cron.Name, "test_name")
	assert.Equal(t, cron.CronExpression, "* * * * * *")
	assert.Equal(t, cron.WorkflowSpec, "workflow")
}
