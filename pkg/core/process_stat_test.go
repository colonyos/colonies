package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessStat(t *testing.T) {
	stat := CreateProcessStat(1, 2, 3, 4)
	jsonString, err := stat.ToJSON()
	assert.Nil(t, err)

	stat2, err := ConvertJSONToProcessStat(jsonString)
	assert.Nil(t, err)
	assert.True(t, stat.Equals(stat2))
}
