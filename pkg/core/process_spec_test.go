package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessSpecJSON(t *testing.T) {
	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1
	in := make(map[string]string)
	in["test_key"] = "test_value"
	processSpec := CreateProcessSpec(colonyID, []string{}, runtimeType, timeout, maxRetries, mem, cores, gpus, in)

	jsonString, err := processSpec.ToJSON()
	assert.Nil(t, err)

	processSpec2, err := ConvertJSONToProcessSpec(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, processSpec.TargetColonyID, processSpec2.TargetColonyID)
	assert.Equal(t, processSpec.TargetRuntimeIDs, processSpec2.TargetRuntimeIDs)
	assert.Equal(t, processSpec.RuntimeType, processSpec2.RuntimeType)
	assert.Equal(t, processSpec.Timeout, processSpec2.Timeout)
	assert.Equal(t, processSpec.MaxRetries, processSpec2.MaxRetries)
	assert.Equal(t, processSpec.Cores, processSpec2.Cores)
	assert.Equal(t, processSpec.GPUs, processSpec2.GPUs)
	assert.Equal(t, processSpec.In, processSpec2.In)
}
