package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEmptyProcessSpe(t *testing.T) {
	processSpec := CreateEmptyProcessSpec()
	assert.NotNil(t, processSpec)
}

func TestProcessSpecJSON(t *testing.T) {
	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	maxExecTime := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1
	env := make(map[string]string)
	env["test_key"] = "test_value"

	processSpec := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, maxExecTime, maxRetries, mem, cores, gpus, env, []string{"test_name2"}, 5)

	jsonString, err := processSpec.ToJSON()
	assert.Nil(t, err)

	processSpec2, err := ConvertJSONToProcessSpec(jsonString + "error")
	assert.NotNil(t, err)

	processSpec2, err = ConvertJSONToProcessSpec(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, processSpec.Conditions.ColonyID, processSpec2.Conditions.ColonyID)
	assert.Equal(t, processSpec.MaxExecTime, processSpec2.MaxExecTime)
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

func TestProcessSpecEquals(t *testing.T) {
	colonyID := GenerateRandomID()
	runtimeType := "test_runtime_type"
	runtime1ID := GenerateRandomID()
	runtime2ID := GenerateRandomID()
	runtime3ID := GenerateRandomID()
	maxExecTime := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1
	env := make(map[string]string)
	env["test_key"] = "test_value"

	env2 := make(map[string]string)
	env2["test_key2"] = "test_value2"

	processSpec1 := CreateProcessSpec("test_name", "test_image", "test_cmd", []string{"test_arg"}, []string{"test_volumes"}, []string{"test_ports"}, colonyID, []string{runtime1ID, runtime2ID}, runtimeType, maxExecTime, maxRetries, mem, cores, gpus, env, []string{}, 1)

	processSpec2 := CreateProcessSpec("test_name", "test_image2", "test_cmd2", []string{"test_arg2"}, []string{"test_volumes2"}, []string{"test_ports2"}, colonyID, []string{runtime3ID}, runtimeType+"2", 3, 100, 200, 4, 2, env2, []string{}, 1)

	assert.True(t, processSpec1.Equals(processSpec1))
	assert.False(t, processSpec1.Equals(nil))
	assert.False(t, processSpec1.Equals(processSpec2))
}
