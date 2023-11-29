package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFunction(t *testing.T) {
	functionID := GenerateRandomID()
	executorName := GenerateRandomID()
	colonyName := GenerateRandomID()

	function1 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.True(t, function1.Equals(&function1))

	function2 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  2,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function2))

	function3 := Function{
		FunctionID:   functionID + "bla",
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function3))

	function4 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName + "bla",
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function4))

	function5 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName + "bla",
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function5))

	function6 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1" + "bla",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function6))

	function7 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc2",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function7))

	function8 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		AvgWaitTime:  2.1,
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function8))

	function9 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  1.1}
	assert.False(t, function1.Equals(&function9))

	function10 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      2,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function10))

	function11 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  2,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function11))

	function12 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  2,
		MinExecTime:  1,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function12))

	function13 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  2,
		MaxExecTime:  1,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function13))

	function14 := Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ColonyName:   colonyName,
		FuncName:     "testfunc1",
		MinWaitTime:  1,
		MaxWaitTime:  1,
		MinExecTime:  1,
		MaxExecTime:  2,
		Counter:      1,
		AvgWaitTime:  1.1,
		AvgExecTime:  0.1}
	assert.False(t, function1.Equals(&function14))
}

func TestFunctionToJSON(t *testing.T) {
	function1 := Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	jsonStr, err := function1.ToJSON()
	assert.Nil(t, err)

	function2, err := ConvertJSONToFunction(jsonStr)
	assert.Nil(t, err)
	assert.True(t, function1.Equals(function2))
}

func TestIsFunctionArraysEqual(t *testing.T) {
	function1 := &Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	function2 := &Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	functions1 := []*Function{function1, function2}
	assert.True(t, IsFunctionArraysEqual(functions1, functions1))

	function3 := &Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	function4 := &Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc3", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	functions2 := []*Function{function3, function4}
	assert.False(t, IsFunctionArraysEqual(functions1, functions2))
}

func TestFunctionsToJSON(t *testing.T) {
	function1 := &Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc1", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	function2 := &Function{FunctionID: GenerateRandomID(), ExecutorName: GenerateRandomID(), ColonyName: GenerateRandomID(), FuncName: "testfunc2", AvgWaitTime: 1.1, AvgExecTime: 0.1}
	functions1 := []*Function{function1, function2}

	jsonStr, err := ConvertFunctionArrayToJSON(functions1)
	assert.Nil(t, err)

	functions2, err := ConvertJSONToFunctionArray(jsonStr)
	assert.Nil(t, err)

	assert.True(t, IsFunctionArraysEqual(functions1, functions2))
}
