package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFunction(t *testing.T) {
	functionID := GenerateRandomID()
	executorID := GenerateRandomID()
	colonyID := GenerateRandomID()

	function1 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.True(t, function1.Equals(&function1))

	function2 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1"}}
	assert.False(t, function1.Equals(&function2))

	function3 := Function{
		FunctionID:  functionID + "bla",
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function3))

	function4 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID + "bla",
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function4))

	function5 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID + "bla",
		FuncName:    "testfunc1",
		Desc:        "unit test function 2",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function5))

	function6 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1" + "bla",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function6))

	function7 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc2",
		Desc:        "unit test function" + "bla",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function7))

	function8 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		AvgWaitTime: 2.1,
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function8))

	function9 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 1.1, Args: []string{}}
	assert.False(t, function1.Equals(&function9))

	function10 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     2,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function10))

	function11 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 2,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function11))

	function12 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 2,
		MinExecTime: 1,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function12))

	function13 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 2,
		MaxExecTime: 1,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function13))

	function14 := Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		FuncName:    "testfunc1",
		Desc:        "unit test function",
		MinWaitTime: 1,
		MaxWaitTime: 1,
		MinExecTime: 1,
		MaxExecTime: 2,
		Counter:     1,
		AvgWaitTime: 1.1,
		AvgExecTime: 0.1,
		Args:        []string{"arg1", "arg2"}}
	assert.False(t, function1.Equals(&function14))
}

func TestFunctionToJSON(t *testing.T) {
	function1 := Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1", "arg2"}}
	jsonStr, err := function1.ToJSON()
	assert.Nil(t, err)

	function2, err := ConvertJSONToFunction(jsonStr)
	assert.Nil(t, err)
	assert.True(t, function1.Equals(function2))
}

func TestIsFunctionArraysEqual(t *testing.T) {
	function1 := &Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}
	function2 := &Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc2", Desc: "unit test function 2", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1", "arg2", "arg3"}}
	functions1 := []*Function{function1, function2}
	assert.True(t, IsFunctionArraysEqual(functions1, functions1))

	function3 := &Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}
	function4 := &Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc3", Desc: "unit test function 2", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1", "arg2", "arg3"}}
	functions2 := []*Function{function3, function4}
	assert.False(t, IsFunctionArraysEqual(functions1, functions2))
}

func TestFunctionsToJSON(t *testing.T) {
	function1 := &Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc1", Desc: "unit test function", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1"}}
	function2 := &Function{FunctionID: GenerateRandomID(), ExecutorID: GenerateRandomID(), ColonyID: GenerateRandomID(), FuncName: "testfunc2", Desc: "unit test function 2", AvgWaitTime: 1.1, AvgExecTime: 0.1, Args: []string{"arg1", "arg2", "arg3"}}
	functions1 := []*Function{function1, function2}

	jsonStr, err := ConvertFunctionArrayToJSON(functions1)
	assert.Nil(t, err)

	functions2, err := ConvertJSONToFunctionArray(jsonStr)
	assert.Nil(t, err)

	assert.True(t, IsFunctionArraysEqual(functions1, functions2))
}
