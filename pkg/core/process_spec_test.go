package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessSpecJSON(t *testing.T) {
	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1
	in := make(map[string]string)
	in["test_key"] = "test_value"
	processSpec := CreateProcessSpec(colonyID, []string{runtime1ID, runtime2ID}, runtimeType, timeout, maxRetries, mem, cores, gpus, in)

	jsonString, err := processSpec.ToJSON()
	assert.Nil(t, err)

	processSpec2, err := ConvertJSONToProcessSpec(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, processSpec.Conditions.ColonyID, processSpec2.Conditions.ColonyID)
	assert.Equal(t, processSpec.Timeout, processSpec2.Timeout)
	assert.Equal(t, processSpec.MaxRetries, processSpec2.MaxRetries)
	assert.Equal(t, processSpec.Conditions.RuntimeIDs, processSpec2.Conditions.RuntimeIDs)
	assert.Contains(t, processSpec.Conditions.RuntimeIDs, runtime1ID)
	assert.Contains(t, processSpec.Conditions.RuntimeIDs, runtime2ID)
	assert.Equal(t, processSpec.Conditions.RuntimeType, processSpec2.Conditions.RuntimeType)
	assert.Equal(t, processSpec.Conditions.Mem, processSpec2.Conditions.Mem)
	assert.Equal(t, processSpec.Conditions.Cores, processSpec2.Conditions.Cores)
	assert.Equal(t, processSpec.Conditions.GPUs, processSpec2.Conditions.GPUs)
	assert.Equal(t, processSpec.Env, processSpec2.Env)
}
