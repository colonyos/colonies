package kvstore

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestFunctionClosedDB(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)

	db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		Counter:      2,
		MinWaitTime:  1.0,
		MaxWaitTime:  2.0,
		MinExecTime:  3.0,
		MaxExecTime:  4.0,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	// KVStore operations work even after close (in-memory store)
	err = db.AddFunction(function1)
	assert.Nil(t, err)

	_, err = db.GetFunctionByID("invalid_id")
	assert.Nil(t, err) // Returns nil for non-existing

	_, err = db.GetFunctionsByExecutorName(colonyName, "invalid_id")
	assert.Nil(t, err) // Returns empty slice

	_, err = db.GetFunctionsByColonyName("invalid_name")
	assert.Nil(t, err) // Returns empty slice

	_, err = db.GetFunctionsByExecutorAndName(colonyName, "invalid_id", "invalid_name")
	assert.Nil(t, err) // Returns nil for non-existing

	err = db.UpdateFunctionStats(colonyName, "invalid_id", "invalid_name", 20, 0.1, 0.2, 0.3, 0.4, 2.0, 2.1)
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveFunctionByID("invalid_id")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveFunctionByName(colonyName, "invalid_id", "invalid_name")
	assert.NotNil(t, err) // Should error for non-existing

	err = db.RemoveFunctionsByExecutorName(colonyName, "invalid_id")
	assert.Nil(t, err) // No error when nothing to remove

	err = db.RemoveFunctionsByColonyName("invalid_name")
	assert.Nil(t, err) // No error when nothing to remove

	err = db.RemoveFunctions()
	assert.Nil(t, err)
}

func TestAddFunction(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   core.GenerateRandomID(),
		FuncName:     "testfunc1",
		Counter:      2,
		MinWaitTime:  1.0,
		MaxWaitTime:  2.0,
		MinExecTime:  3.0,
		MaxExecTime:  4.0,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	assert.True(t, function1.Equals(functions[0]))

	// Test adding function with same ID (should update)
	function1.Counter = 10
	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)
	assert.Equal(t, functions[0].Counter, 10)
}

func TestGetFunctionByExecutorIDAndName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   core.GenerateRandomID(),
		FuncName:     "testfunc1",
		Counter:      2,
		MinWaitTime:  1.0,
		MaxWaitTime:  2.0,
		MinExecTime:  3.0,
		MaxExecTime:  4.0,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functionFromDB, err := db.GetFunctionsByExecutorAndName(function1.ColonyName, function1.ExecutorName, function1.FuncName)
	assert.Nil(t, err)
	assert.True(t, function1.Equals(functionFromDB))

	// Test non-existing function
	functionFromDB, err = db.GetFunctionsByExecutorAndName(function1.ColonyName, function1.ExecutorName, "does_not_exists")
	assert.Nil(t, err)
	assert.Nil(t, functionFromDB)

	// Test non-existing executor
	functionFromDB, err = db.GetFunctionsByExecutorAndName(function1.ColonyName, "invalid_executor", function1.FuncName)
	assert.Nil(t, err)
	assert.Nil(t, functionFromDB)

	// Test non-existing colony
	functionFromDB, err = db.GetFunctionsByExecutorAndName("invalid_colony", function1.ExecutorName, function1.FuncName)
	assert.Nil(t, err)
	assert.Nil(t, functionFromDB)
}

func TestGetFunctionByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		Counter:      3,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2, err := db.GetFunctionByID(function1.FunctionID)
	assert.Nil(t, err)

	assert.True(t, function1.Equals(function2))

	// Test non-existing ID
	function3, err := db.GetFunctionByID("non_existing_id")
	assert.Nil(t, err)
	assert.Nil(t, function3)
}

func TestGetFunctionByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	// Add function from different colony
	otherColony := core.GenerateRandomID()
	function3 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   otherColony,
		FuncName:     "testfunc3",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function3)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Nil(t, err)

	assert.Len(t, functions, 2)

	// Verify correct functions returned
	foundFunc1, foundFunc2 := false, false
	for _, fn := range functions {
		if fn.FunctionID == function1.FunctionID {
			foundFunc1 = true
		}
		if fn.FunctionID == function2.FunctionID {
			foundFunc2 = true
		}
		assert.Equal(t, fn.ColonyName, colonyName)
	}
	assert.True(t, foundFunc1 && foundFunc2)

	// Test non-existing colony
	functionsEmpty, err := db.GetFunctionsByColonyName("non_existing_colony")
	assert.Nil(t, err)
	assert.Empty(t, functionsEmpty)
}

func TestGetFunctionsByExecutorName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	// Add function from different executor
	otherExecutor := core.GenerateRandomID()
	function3 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: otherExecutor,
		ColonyName:   colonyName,
		FuncName:     "testfunc3",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function3)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByExecutorName(colonyName, executorName)
	assert.Nil(t, err)
	assert.Len(t, functions, 2)

	// Verify correct functions returned
	for _, fn := range functions {
		assert.Equal(t, fn.ExecutorName, executorName)
		assert.Equal(t, fn.ColonyName, colonyName)
	}

	// Test non-existing executor
	functionsEmpty, err := db.GetFunctionsByExecutorName(colonyName, "non_existing_executor")
	assert.Nil(t, err)
	assert.Empty(t, functionsEmpty)
}

func TestUpdateFunctionStats(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		Counter:      10,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	assert.Equal(t, function1.Counter, 10)
	assert.Equal(t, function1.AvgWaitTime, 1.1)
	assert.Equal(t, function1.AvgExecTime, 0.1)

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	err = db.UpdateFunctionStats(function1.ColonyName, function1.ExecutorName, function1.FuncName, 20, 0.1, 0.2, 0.3, 0.4, 2.0, 2.1)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	assert.Equal(t, functions[0].Counter, 20)
	assert.Equal(t, functions[0].MinWaitTime, 0.1)
	assert.Equal(t, functions[0].MaxWaitTime, 0.2)
	assert.Equal(t, functions[0].MinExecTime, 0.3)
	assert.Equal(t, functions[0].MaxExecTime, 0.4)
	assert.Equal(t, functions[0].AvgWaitTime, 2.0)
	assert.Equal(t, functions[0].AvgExecTime, 2.1)

	// Test updating non-existing function
	err = db.UpdateFunctionStats("invalid_colony", "invalid_executor", "invalid_function", 30, 0.5, 0.6, 0.7, 0.8, 3.0, 3.1)
	assert.NotNil(t, err)
}

func TestRemoveFunctionByExecutorID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 2)

	err = db.RemoveFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.Equal(t, functions[0].FunctionID, function2.FunctionID)

	// Test removing from non-existing executor
	err = db.RemoveFunctionsByExecutorName(colonyName, "non_existing_executor")
	assert.Nil(t, err) // Should not error
}

func TestRemoveFunctionByID(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 2)

	err = db.RemoveFunctionByID(function1.FunctionID)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(function2))

	// Test removing non-existing function
	err = db.RemoveFunctionByID("non_existing_id")
	assert.NotNil(t, err)
}

func TestRemoveFunctionByName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 2)

	err = db.RemoveFunctionByName(function1.ColonyName, function1.ExecutorName, "testfunc1")
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(function2))

	// Test removing non-existing function
	err = db.RemoveFunctionByName(colonyName, executorName, "non_existing_function")
	assert.NotNil(t, err)
}

func TestRemoveFunctionByColonyName(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName1,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName1,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	function3 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName2,
		FuncName:     "testfunc3",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function3)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName1)
	assert.Len(t, functions, 2)

	functions, err = db.GetFunctionsByColonyName(colonyName2)
	assert.Len(t, functions, 1)

	err = db.RemoveFunctionsByColonyName(function1.ColonyName)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName1)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyName(colonyName2)
	assert.Len(t, functions, 1)

	// Test removing from non-existing colony
	err = db.RemoveFunctionsByColonyName("non_existing_colony")
	assert.Nil(t, err) // Should not error
}

func TestRemoveFunctions(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName1,
		FuncName:     "testfunc1",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName1,
		FuncName:     "testfunc2",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	function3 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: core.GenerateRandomID(),
		ColonyName:   colonyName2,
		FuncName:     "testfunc3",
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1,
	}

	err = db.AddFunction(function3)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName1)
	assert.Len(t, functions, 2)

	functions, err = db.GetFunctionsByColonyName(colonyName2)
	assert.Len(t, functions, 1)

	err = db.RemoveFunctions()
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName1)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyName(colonyName2)
	assert.Len(t, functions, 0)
}

func TestFunctionComplexScenarios(t *testing.T) {
	db := NewKVStoreDatabase()
	err := db.Initialize()
	assert.Nil(t, err)
	defer db.Close()

	// Test function with complex statistics
	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	function := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "complex_function",
		Counter:      100,
		MinWaitTime:  0.5,
		MaxWaitTime:  10.0,
		MinExecTime:  1.0,
		MaxExecTime:  30.0,
		AvgWaitTime:  2.5,
		AvgExecTime:  5.2,
	}

	err = db.AddFunction(function)
	assert.Nil(t, err)

	// Test multiple updates
	for i := 0; i < 5; i++ {
		err = db.UpdateFunctionStats(colonyName, executorName, "complex_function",
			function.Counter+i+1, 0.1, 15.0, 0.5, 40.0, 3.0, 6.0)
		assert.Nil(t, err)
	}

	// Verify final state
	updatedFunction, err := db.GetFunctionsByExecutorAndName(colonyName, executorName, "complex_function")
	assert.Nil(t, err)
	assert.NotNil(t, updatedFunction)
	assert.Equal(t, updatedFunction.Counter, 105) // 100 + 5
	assert.Equal(t, updatedFunction.AvgWaitTime, 3.0)
	assert.Equal(t, updatedFunction.AvgExecTime, 6.0)

	// Test multiple functions with same name in different executors
	otherExecutor := core.GenerateRandomID()
	function2 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: otherExecutor,
		ColonyName:   colonyName,
		FuncName:     "complex_function", // Same name, different executor
		Counter:      50,
		AvgWaitTime:  1.0,
		AvgExecTime:  2.0,
	}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	// Both should be retrievable independently
	func1, err := db.GetFunctionsByExecutorAndName(colonyName, executorName, "complex_function")
	assert.Nil(t, err)
	assert.Equal(t, func1.ExecutorName, executorName)

	func2, err := db.GetFunctionsByExecutorAndName(colonyName, otherExecutor, "complex_function")
	assert.Nil(t, err)
	assert.Equal(t, func2.ExecutorName, otherExecutor)
}