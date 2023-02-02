package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatistics(t *testing.T) {
	stat := CreateStatistics(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	jsonString, err := stat.ToJSON()
	assert.Nil(t, err)

	stat2, err := ConvertJSONToStatistics(jsonString + "error")
	assert.NotNil(t, err)

	stat2, err = ConvertJSONToStatistics(jsonString)
	assert.Nil(t, err)
	assert.True(t, stat.Equals(stat2))

	assert.True(t, stat2.Colonies == 1)
	assert.True(t, stat2.Executors == 2)
	assert.True(t, stat2.WaitingProcesses == 3)
	assert.True(t, stat2.RunningProcesses == 4)
	assert.True(t, stat2.SuccessfulProcesses == 5)
	assert.True(t, stat2.FailedProcesses == 6)
	assert.True(t, stat2.WaitingWorkflows == 7)
	assert.True(t, stat2.RunningWorkflows == 8)
	assert.True(t, stat2.SuccessfulWorkflows == 9)
	assert.True(t, stat2.FailedWorkflows == 10)
}

func TestStatisticsEquals(t *testing.T) {
	stat := CreateStatistics(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	assert.True(t, stat.Equals(stat))
	assert.False(t, stat.Equals(nil))
}
