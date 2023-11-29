package server

import (
	"fmt"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFunctionSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 and executor3 are member of colony1
	//   executor2 is member of colony2

	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(env.colony1Name, executor3.Name, env.colony1PrvKey)
	assert.Nil(t, err)

	function1 := &core.Function{ExecutorName: env.executor1Name, ColonyName: env.colony1Name, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	function2 := &core.Function{ExecutorName: env.executor2Name, ColonyName: env.colony1Name, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	function3 := &core.Function{ExecutorName: env.executor1Name, ColonyName: env.colony2Name, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err = client.AddFunction(function1, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function1, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function2, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function3, env.executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function1, executor3PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function1, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetFunctionsSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	function1 := &core.Function{ExecutorName: env.executor1Name, ColonyName: env.colony1Name, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	function2 := &core.Function{ExecutorName: env.executor1Name, ColonyName: env.colony1Name, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	_, err := client.AddFunction(function1, env.executor1PrvKey)
	assert.Nil(t, err)
	_, err = client.AddFunction(function2, env.executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFunctionsByExecutorName(env.colony1Name, env.executor1Name, env.executor2PrvKey)
	fmt.Println(err)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFunctionsByExecutorName(env.colony1Name, env.executor1Name, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFunctionsByExecutorName(env.colony1Name, env.executor1Name, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFunctionsByExecutorName(env.colony1Name, env.executor1Name, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestDeleteFunctionSecurity(t *testing.T) {
	env, client, server, _, done := setupTestEnv1(t)

	// The setup looks like this:
	//   executor1 and exector 3 are member of colony1
	//   executor2 is member of colony2

	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(env.colony1Name, executor3.Name, env.colony1PrvKey)
	assert.Nil(t, err)

	function := &core.Function{ExecutorName: env.executor1Name, ColonyName: env.colony1Name, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	addedFunction, err := client.AddFunction(function, env.executor1PrvKey)
	assert.Nil(t, err)

	err = client.DeleteFunction(addedFunction.FunctionID, env.executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFunction(addedFunction.FunctionID, env.colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFunction(addedFunction.FunctionID, env.colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.DeleteFunction(addedFunction.FunctionID, executor3PrvKey)
	assert.NotNil(t, err)

	err = client.DeleteFunction(addedFunction.FunctionID, env.executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
