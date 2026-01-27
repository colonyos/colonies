package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestFunctionClosedDB(t *testing.T) {
	db, err := PrepareTests()
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
		AvgExecTime:  0.1}

	err = db.AddFunction(function1)
	assert.NotNil(t, err)

	_, err = db.GetFunctionByID("invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetFunctionsByExecutorName(colonyName, "invalid_id")
	assert.NotNil(t, err)

	_, err = db.GetFunctionsByColonyName("invalid_name")
	assert.NotNil(t, err)

	_, err = db.GetFunctionsByExecutorAndName(colonyName, "invalid_id", "invalid_name")
	assert.NotNil(t, err)

	err = db.UpdateFunctionStats(colonyName, "invalid_id", "invalid_name", 20, 0.1, 0.2, 0.3, 0.4, 2.0, 2.1)
	assert.NotNil(t, err)

	err = db.RemoveFunctionByID("invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveFunctionByName(colonyName, "invalid_id", "invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveFunctionsByExecutorName(colonyName, "invalid_id")
	assert.NotNil(t, err)

	err = db.RemoveFunctionsByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.RemoveFunctions()
	assert.NotNil(t, err)
}

func TestAddFunction(t *testing.T) {
	db, err := PrepareTests()
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
		AvgExecTime:  0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	assert.True(t, function1.Equals(functions[0]))
}

func TestGetFunctionByExecutorIDAndName(t *testing.T) {
	db, err := PrepareTests()
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
		AvgExecTime:  0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functionFromDB, err := db.GetFunctionsByExecutorAndName(function1.ColonyName, function1.ExecutorName, function1.FuncName)
	assert.Nil(t, err)
	assert.True(t, function1.Equals(functionFromDB))

	functionFromDB, err = db.GetFunctionsByExecutorAndName(function1.ColonyName, function1.ExecutorName, "does_not_exists")
	assert.Nil(t, err)
	assert.Nil(t, functionFromDB)
}

func TestGetFunctionByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName, FuncName: "testfunc1", Counter: 3, AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2, err := db.GetFunctionByID(function1.FunctionID)
	assert.Nil(t, err)

	assert.True(t, function1.Equals(function2))
}

func TestGetFunctionByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Nil(t, err)

	assert.Len(t, functions, 2)
}

func TestUpdateFunctionStats(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName, FuncName: "testfunc1", Counter: 10, AvgWaitTime: 1.1, AvgExecTime: 0.1}

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
}

func TestRemoveFunctionByExecutorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 2)

	err = db.RemoveFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
}

func TestRemoveFunctionByID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executorName, ColonyName: colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executorName, ColonyName: colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 2)

	err = db.RemoveFunctionByID(function1.FunctionID)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(function2))
}

func TestRemoveFunctionByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executorName, ColonyName: colonyName, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: executorName, ColonyName: colonyName, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 2)

	err = db.RemoveFunctionByName(function1.ColonyName, function1.ExecutorName, "testfunc1")
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(function2))
}

func TestRemoveFunctionByColonyName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	function3 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName2, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}

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
}

func TestFunctionWithDescriptionAndArgs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	// Create function with description and args
	args := []*core.FunctionArg{
		{Name: "query", Type: "string", Description: "Search query", Required: true},
		{Name: "limit", Type: "integer", Description: "Max results", Required: false},
		{Name: "format", Type: "string", Description: "Output format", Enum: []string{"json", "text", "xml"}},
	}

	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "search_tool",
		Description:  "Search for content in the database",
		Args:         args,
		Counter:      0,
		MinWaitTime:  0.0,
		MaxWaitTime:  0.0,
		MinExecTime:  0.0,
		MaxExecTime:  0.0,
		AvgWaitTime:  0.0,
		AvgExecTime:  0.0,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	// Retrieve and verify
	functionFromDB, err := db.GetFunctionByID(function1.FunctionID)
	assert.Nil(t, err)
	assert.NotNil(t, functionFromDB)

	assert.Equal(t, function1.Description, functionFromDB.Description)
	assert.Equal(t, len(function1.Args), len(functionFromDB.Args))

	// Verify each arg
	for i, arg := range function1.Args {
		assert.Equal(t, arg.Name, functionFromDB.Args[i].Name)
		assert.Equal(t, arg.Type, functionFromDB.Args[i].Type)
		assert.Equal(t, arg.Description, functionFromDB.Args[i].Description)
		assert.Equal(t, arg.Required, functionFromDB.Args[i].Required)
		assert.Equal(t, len(arg.Enum), len(functionFromDB.Args[i].Enum))
	}
}

func TestFunctionWithEmptyDescriptionAndArgs(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName := core.GenerateRandomID()
	executorName := core.GenerateRandomID()

	// Function without description and args (backwards compatibility)
	function1 := &core.Function{
		FunctionID:   core.GenerateRandomID(),
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "simple_func",
		Counter:      5,
		AvgWaitTime:  1.0,
		AvgExecTime:  2.0,
	}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functionFromDB, err := db.GetFunctionByID(function1.FunctionID)
	assert.Nil(t, err)
	assert.NotNil(t, functionFromDB)

	assert.Equal(t, "", functionFromDB.Description)
	assert.Nil(t, functionFromDB.Args)
}

func TestRemoveFunctions(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyName1 := core.GenerateRandomID()
	colonyName2 := core.GenerateRandomID()

	function1 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName1, FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	function3 := &core.Function{FunctionID: core.GenerateRandomID(), ExecutorName: core.GenerateRandomID(), ColonyName: colonyName2, FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}

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
