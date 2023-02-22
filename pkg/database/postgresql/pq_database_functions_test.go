package postgresql

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestAddFunction(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: core.GenerateRandomID(), Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByExecutorID(function1.ExecutorID)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	assert.True(t, function1.Equals(&functions[0]))
}

func TestGetFunctionByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colonyID)
	assert.Nil(t, err)

	assert.Len(t, functions, 2)
}

func TestUpdateFunctionTimes(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	err = db.UpdateFunctionTimes(function1.ExecutorID, function1.Name, 2.0, 2.1)
	assert.Nil(t, err)

	assert.Equal(t, function1.AvgWaitTime, 1.1)
	assert.Equal(t, function1.AvgExecTime, 0.1)

	functions, err := db.GetFunctionsByExecutorID(function1.ExecutorID)
	assert.Nil(t, err)
	assert.Len(t, functions, 1)

	assert.Equal(t, functions[0].AvgWaitTime, 2.0)
	assert.Equal(t, functions[0].AvgExecTime, 2.1)
}

func TestDeleteFunctionByExecutorID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID, Name: "testfunc2", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colonyID)
	assert.Len(t, functions, 2)

	err = db.DeleteFunctionsByExecutorID(function1.ExecutorID)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyID(colonyID)
	assert.Len(t, functions, 1)
}

func TestDeleteFunctionByName(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID := core.GenerateRandomID()
	executorID := core.GenerateRandomID()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executorID, ColonyID: colonyID, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: executorID, ColonyID: colonyID, Name: "testfunc2", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colonyID)
	assert.Len(t, functions, 2)

	err = db.DeleteFunctionByName(function1.ExecutorID, "testfunc1")
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyID(colonyID)
	assert.Len(t, functions, 1)
	assert.True(t, functions[0].Equals(&function2))
}

func TestDeleteFunctionByColonyID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID1 := core.GenerateRandomID()
	colonyID2 := core.GenerateRandomID()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID1, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID1, Name: "testfunc2", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	function3 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID2, Name: "testfunc3", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function3)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colonyID1)
	assert.Len(t, functions, 2)

	functions, err = db.GetFunctionsByColonyID(colonyID2)
	assert.Len(t, functions, 1)

	err = db.DeleteFunctionsByColonyID(function1.ColonyID)
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyID(colonyID1)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyID(colonyID2)
	assert.Len(t, functions, 1)
}

func TestDeleteFunctions(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	defer db.Close()

	colonyID1 := core.GenerateRandomID()
	colonyID2 := core.GenerateRandomID()

	function1 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID1, Name: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function1)
	assert.Nil(t, err)

	function2 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID1, Name: "testfunc2", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function2)
	assert.Nil(t, err)

	function3 := core.Function{FunctionID: core.GenerateRandomID(), ExecutorID: core.GenerateRandomID(), ColonyID: colonyID2, Name: "testfunc3", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}

	err = db.AddFunction(function3)
	assert.Nil(t, err)

	functions, err := db.GetFunctionsByColonyID(colonyID1)
	assert.Len(t, functions, 2)

	functions, err = db.GetFunctionsByColonyID(colonyID2)
	assert.Len(t, functions, 1)

	err = db.DeleteFunctions()
	assert.Nil(t, err)

	functions, err = db.GetFunctionsByColonyID(colonyID1)
	assert.Len(t, functions, 0)

	functions, err = db.GetFunctionsByColonyID(colonyID2)
	assert.Len(t, functions, 0)
}
