package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessStat(t *testing.T) {
	stat := CreateProcessStat(1, 2, 3, 4)
	jsonString, err := stat.ToJSON()
	assert.Nil(t, err)

	stat2, err := ConvertJSONToProcessStat(jsonString + "error")
	assert.NotNil(t, err)

	stat2, err = ConvertJSONToProcessStat(jsonString)
	assert.Nil(t, err)
	assert.True(t, stat.Equals(stat2))
}

func TestProcessStatEquals(t *testing.T) {
	stat := CreateProcessStat(1, 2, 3, 4)

	assert.True(t, stat.Equals(stat))
	assert.False(t, stat.Equals(nil))
}
