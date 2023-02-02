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
	executorType := "test_executor_type"
	executor1ID := GenerateRandomID()
	executor2ID := GenerateRandomID()
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3
	env := make(map[string]string)
	env["test_key"] = "test_value"

	processSpec := CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, env, []string{"test_name2"}, 5)

	jsonString, err := processSpec.ToJSON()
	assert.Nil(t, err)

	processSpec2, err := ConvertJSONToProcessSpec(jsonString + "error")
	assert.NotNil(t, err)

	processSpec2, err = ConvertJSONToProcessSpec(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, processSpec.Conditions.ColonyID, processSpec2.Conditions.ColonyID)
	assert.Equal(t, processSpec.MaxExecTime, processSpec2.MaxExecTime)
	assert.Equal(t, processSpec.MaxRetries, processSpec2.MaxRetries)
	assert.Equal(t, processSpec.Conditions.ExecutorIDs, processSpec2.Conditions.ExecutorIDs)
	assert.Contains(t, processSpec.Conditions.ExecutorIDs, executor1ID)
	assert.Contains(t, processSpec.Conditions.ExecutorIDs, executor2ID)
	assert.Equal(t, processSpec.Conditions.ExecutorType, processSpec2.Conditions.ExecutorType)
	assert.Equal(t, processSpec.Env, processSpec2.Env)
}

func TestProcessSpecEquals(t *testing.T) {
	colonyID := GenerateRandomID()
	executorType := "test_executor_type"
	executor1ID := GenerateRandomID()
	executor2ID := GenerateRandomID()
	executor3ID := GenerateRandomID()
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3
	env := make(map[string]string)
	env["test_key"] = "test_value"

	env2 := make(map[string]string)
	env2["test_key2"] = "test_value2"

	processSpec1 := CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{executor1ID, executor2ID}, executorType, maxWaitTime, maxExecTime, maxRetries, env, []string{}, 1)

	processSpec2 := CreateProcessSpec("test_name", "test_func", []string{"test_arg2"}, colonyID, []string{executor3ID}, executorType+"2", 200, 4, 2, env2, []string{}, 1)

	assert.True(t, processSpec1.Equals(processSpec1))
	assert.False(t, processSpec1.Equals(nil))
	assert.False(t, processSpec1.Equals(processSpec2))
}
