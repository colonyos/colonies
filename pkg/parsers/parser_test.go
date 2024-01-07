package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertCPUToInt(t *testing.T) {
	cpu := "100m"
	result, err := ConvertCPUToInt(cpu)
	assert.Nil(t, err)
	assert.Equal(t, int64(100), result)

	cpu = "10"
	result, err = ConvertCPUToInt(cpu)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), result)

	cpu = "m"
	result, err = ConvertCPUToInt(cpu)
	assert.NotNil(t, err)
}

func TestConvertMemoryToInt(t *testing.T) {
	cases := []struct {
		mem      string
		expected int64
	}{
		{"1Gi", 1073741824},
		{"1G", 1000000000},
		{"1M", 1000000},
		{"1K", 1000},
		{"1", -1},
		{"1Mi", 1048576},
		{"2GiB", 2 * 1073741824},
		{"", 0},      // Test empty string
		{"1XYZ", -1}, // Test invalid unit
	}

	for _, tc := range cases {
		result, err := ConvertMemoryToBytes(tc.mem)
		if tc.expected == -1 {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		}
	}
}
