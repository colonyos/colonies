package function_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestAddFunctionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 and executor3 are member of colony1
	//   executor2 is member of colony2

	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	function1 := core.CreateFunction(core.GenerateRandomID(), env.Executor1Name, "test_executor_type", env.Colony1Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	function2 := core.CreateFunction(core.GenerateRandomID(), env.Executor2Name, "test_executor_type", env.Colony1Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	function3 := core.CreateFunction(core.GenerateRandomID(), env.Executor1Name, "test_executor_type", env.Colony2Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err = client.AddFunction(function1, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function1, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function2, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function3, env.Executor1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function1, executor3PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.AddFunction(function1, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestGetFunctionsSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 is member of colony1
	//   executor2 is member of colony2

	function1 := core.CreateFunction(core.GenerateRandomID(), env.Executor1Name, "test_executor_type", env.Colony1Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	function2 := core.CreateFunction(core.GenerateRandomID(), env.Executor1Name, "test_executor_type", env.Colony1Name, "testfunc2", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	_, err := client.AddFunction(function1, env.Executor1PrvKey)
	assert.Nil(t, err)
	_, err = client.AddFunction(function2, env.Executor1PrvKey)
	assert.Nil(t, err)

	_, err = client.GetFunctionsByExecutor(env.Colony1Name, env.Executor1Name, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFunctionsByExecutor(env.Colony1Name, env.Executor1Name, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFunctionsByExecutor(env.Colony1Name, env.Executor1Name, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	_, err = client.GetFunctionsByExecutor(env.Colony1Name, env.Executor1Name, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}

func TestRemoveFunctionSecurity(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv1(t)

	// The setup looks like this:
	//   executor1 and exector 3 are member of colony1
	//   executor2 is member of colony2

	executor3, executor3PrvKey, err := utils.CreateTestExecutorWithKey(env.Colony1Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor3, env.Colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(env.Colony1Name, executor3.Name, env.Colony1PrvKey)
	assert.Nil(t, err)

	function := core.CreateFunction(core.GenerateRandomID(), env.Executor1Name, "test_executor_type", env.Colony1Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)

	addedFunction, err := client.AddFunction(function, env.Executor1PrvKey)
	assert.Nil(t, err)
	
	err = client.RemoveFunction(addedFunction.FunctionID, env.Executor2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFunction(addedFunction.FunctionID, env.Colony1PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFunction(addedFunction.FunctionID, env.Colony2PrvKey)
	assert.NotNil(t, err) // Should not work

	err = client.RemoveFunction(addedFunction.FunctionID, executor3PrvKey)
	assert.NotNil(t, err)

	err = client.RemoveFunction(addedFunction.FunctionID, env.Executor1PrvKey)
	assert.Nil(t, err)

	server.Shutdown()
	<-done
}
