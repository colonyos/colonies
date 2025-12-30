package function_test

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
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

// TestAddFunctionUnauthorized tests adding function from different colony
func TestAddFunctionUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, _, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Try to add function to executor1 with executor2's key
	function := core.CreateFunction(core.GenerateRandomID(), executor1.Name, "test_executor_type", colony1.Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	_, err = client.AddFunction(function, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestAddFunctionToAnotherExecutor tests adding function to another executor
func TestAddFunctionToAnotherExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	executor2, _, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Try to add function to executor2 with executor1's key
	function := core.CreateFunction(core.GenerateRandomID(), executor2.Name, "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	_, err = client.AddFunction(function, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestAddFunctionExecutorNotFound tests adding function for non-existent executor
func TestAddFunctionExecutorNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	// Try to add function to non-existent executor
	function := core.CreateFunction(core.GenerateRandomID(), "non_existent_executor", "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	_, err := client.AddFunction(function, env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestGetFunctionsUnauthorized tests getting functions from different colony
func TestGetFunctionsUnauthorized(t *testing.T) {
	client, s, serverPrvKey, done := server.PrepareTests(t)

	colony1, colonyPrvKey1, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colonyPrvKey1)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony1.Name, executor1.Name, colonyPrvKey1)
	assert.Nil(t, err)

	colony2, colonyPrvKey2, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colonyPrvKey2)
	assert.Nil(t, err)
	err = client.ApproveExecutor(colony2.Name, executor2.Name, colonyPrvKey2)
	assert.Nil(t, err)

	// Add function to executor1
	function := core.CreateFunction(core.GenerateRandomID(), executor1.Name, "test_executor_type", colony1.Name, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	_, err = client.AddFunction(function, executor1PrvKey)
	assert.Nil(t, err)

	// Try to get functions from colony1 with executor2's key
	_, err = client.GetFunctionsByColony(colony1.Name, executor2PrvKey)
	assert.NotNil(t, err)

	s.Shutdown()
	<-done
}

// TestRemoveFunctionNotFound tests removing non-existent function
func TestRemoveFunctionNotFound(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	err := client.RemoveFunction("non_existent_function_id", env.ExecutorPrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}

// TestRemoveFunctionFromAnotherExecutor tests removing function from another executor
func TestRemoveFunctionFromAnotherExecutor(t *testing.T) {
	env, client, server, _, done := server.SetupTestEnv2(t)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(env.ColonyName)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, env.ColonyPrvKey)
	assert.Nil(t, err)
	err = client.ApproveExecutor(env.ColonyName, executor2.Name, env.ColonyPrvKey)
	assert.Nil(t, err)

	// Add function to executor1
	function := core.CreateFunction(core.GenerateRandomID(), env.ExecutorName, "test_executor_type", env.ColonyName, "testfunc1", 0, 0.0, 0.0, 0.0, 0.0, 1.1, 0.1)
	addedFunction, err := client.AddFunction(function, env.ExecutorPrvKey)
	assert.Nil(t, err)

	// Try to remove function from executor1 with executor2's key
	err = client.RemoveFunction(addedFunction.FunctionID, executor2PrvKey)
	assert.NotNil(t, err)

	server.Shutdown()
	<-done
}
