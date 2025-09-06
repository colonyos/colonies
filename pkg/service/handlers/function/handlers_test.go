package function_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/service"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFunction(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	function1 := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	addedFunction1, err := client.AddFunction(function1, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, function1.FuncName, addedFunction1.FuncName)

	_, err = client.AddFunction(function1, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetFunctionsByExecutorID(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	function1 := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err := client.AddFunction(function1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	function2 := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc2", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err = client.AddFunction(function2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	functions, err := client.GetFunctionsByExecutor(env.ColonyName, env.ExecutorName, env.ExecutorPrvKey)
	assert.Nil(t, err)

	counter := 0
	for _, function := range functions {
		if function.FuncName == function1.FuncName || function.FuncName == function2.FuncName {
			counter++
		}
	}

	assert.Equal(t, counter, 2)

	server.Shutdown()
	<-done
}

func TestGetFunctionsByColonyName(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	function1 := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err = client.AddFunction(function1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	function2 := core.CreateFunction(core.GenerateRandomID(), executor2.Name, "test_executor_type", env.ColonyName, "testfunc2", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err = client.AddFunction(function2, executor2PrvKey)
	assert.Nil(t, err)

	functions, err := client.GetFunctionsByColony(env.ColonyName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 2)

	server.Shutdown()
	<-done
}

func TestRemoveFunction(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	function1 := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	addedFunction1, err := client.AddFunction(function1, env.ExecutorPrvKey)
	assert.Nil(t, err)

	function2 := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc2", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err = client.AddFunction(function2, env.ExecutorPrvKey)
	assert.Nil(t, err)

	functions, err := client.GetFunctionsByExecutor(env.ColonyName, env.ExecutorName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 2)

	err = client.RemoveFunction(addedFunction1.FunctionID, env.ExecutorPrvKey)
	assert.Nil(t, err)

	functions, err = client.GetFunctionsByExecutor(env.ColonyName, env.ExecutorName, env.ExecutorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	server.Shutdown()
	<-done
}
