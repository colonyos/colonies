package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFunction(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	function1 := &core.Function{ExecutorName: env.executorName, ColonyName: env.colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	addedFunction1, err := client.AddFunction(function1, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Equal(t, function1.FuncName, addedFunction1.FuncName)

	_, err = client.AddFunction(function1, env.executorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

func TestGetFunctionsByExecutorID(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	function1 := &core.Function{ExecutorName: env.executorName, ColonyName: env.colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err := client.AddFunction(function1, env.executorPrvKey)
	assert.Nil(t, err)

	function2 := &core.Function{ExecutorName: env.executorName, ColonyName: env.colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err = client.AddFunction(function2, env.executorPrvKey)
	assert.Nil(t, err)

	functions, err := client.GetFunctionsByExecutor(env.colonyName, env.executorName, env.executorPrvKey)
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
	env, client, server, _, done := setupTestEnv2(t)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.colonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.colonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.colonyName, executor2.Name, env.colonyPrvKey)
	assert.Nil(t, err)

	function1 := &core.Function{ExecutorName: env.executorName, ColonyName: env.colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err = client.AddFunction(function1, env.executorPrvKey)
	assert.Nil(t, err)

	function2 := &core.Function{ExecutorName: executor2.Name, ColonyName: env.colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err = client.AddFunction(function2, executor2PrvKey)
	assert.Nil(t, err)

	functions, err := client.GetFunctionsByColony(env.colonyName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 2)

	server.Shutdown()
	<-done
}

func TestRemoveFunction(t *testing.T) {
	env, client, server, _, done := setupTestEnv2(t)

	function1 := &core.Function{ExecutorName: env.executorName, ColonyName: env.colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	addedFunction1, err := client.AddFunction(function1, env.executorPrvKey)
	assert.Nil(t, err)

	function2 := &core.Function{ExecutorName: env.executorName, ColonyName: env.colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err = client.AddFunction(function2, env.executorPrvKey)
	assert.Nil(t, err)

	functions, err := client.GetFunctionsByExecutor(env.colonyName, env.executorName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 2)

	err = client.RemoveFunction(addedFunction1.FunctionID, env.executorPrvKey)
	assert.Nil(t, err)

	functions, err = client.GetFunctionsByExecutor(env.colonyName, env.executorName, env.executorPrvKey)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	server.Shutdown()
	<-done
}
