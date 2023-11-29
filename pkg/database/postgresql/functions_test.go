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

	err = db.DeleteFunctionByID("invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteFunctionByName(colonyName, "invalid_id", "invalid_name")
	assert.NotNil(t, err)

	err = db.DeleteFunctionsByExecutorName(colonyName, "invalid_id")
	assert.NotNil(t, err)

	err = db.DeleteFunctionsByColonyName("invalid_name")
	assert.NotNil(t, err)

	err = db.DeleteFunctions()
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

func TestDeleteFunctionByExecutorID(t *testing.T) {
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

	err = db.DeleteFunctionsByExecutorName(function1.ColonyName, function1.ExecutorName)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
}

func TestDeleteFunctionByID(t *testing.T) {
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

	err = db.DeleteFunctionByID(function1.FunctionID)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(function2))
}

func TestDeleteFunctionByName(t *testing.T) {
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

	err = db.DeleteFunctionByName(function1.ColonyName, function1.ExecutorName, "testfunc1")
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(function2))
}

func TestDeleteFunctionByColonyName(t *testing.T) {
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

	err = db.DeleteFunctionsByColonyName(function1.ColonyName)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName1)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyName(colonyName2)
	assert.Len(t, functions, 1)
}

func TestDeleteFunctions(t *testing.T) {
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

	err = db.DeleteFunctions()
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyName(colonyName1)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyName(colonyName2)
	assert.Len(t, functions, 0)
}
