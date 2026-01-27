package core

import "encoding/json"

// FunctionArg describes a function argument/parameter
type FunctionArg struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type Function struct {
	FunctionID   string         `json:"functionid"`
	ExecutorName string         `json:"executorname"`
	ExecutorType string         `json:"executortype"`
	ColonyName   string         `json:"colonyname"`
	FuncName     string         `json:"funcname"`
	Description  string         `json:"description,omitempty"`
	Args         []*FunctionArg `json:"args,omitempty"`
	Counter      int            `json:"counter"`
	MinWaitTime  float64        `json:"minwaittime"`
	MaxWaitTime  float64 `json:"maxwaittime"`
	MinExecTime  float64 `json:"minexectime"`
	MaxExecTime  float64 `json:"maxexectime"`
	AvgWaitTime  float64 `json:"avgwaittime"`
	AvgExecTime  float64 `json:"avgexectime"`
}

func CreateFunction(functionID string,
	executorName string,
	executorType string,
	colonyName string,
	funcName string,
	counter int,
	minWaitTime float64,
	maxWaitTime float64,
	minExecTime float64,
	maxExecTime float64,
	avgWaitTime float64,
	avgExecTime float64) *Function {
	return &Function{
		FunctionID:   functionID,
		ExecutorName: executorName,
		ExecutorType: executorType,
		ColonyName:   colonyName,
		FuncName:     funcName,
		Counter:      counter,
		MinWaitTime:  minWaitTime,
		MaxWaitTime:  maxWaitTime,
		MinExecTime:  minExecTime,
		MaxExecTime:  maxExecTime,
		AvgWaitTime:  avgWaitTime,
		AvgExecTime:  avgExecTime,
	}
}

// CreateFunctionWithDesc creates a Function with description and arguments
func CreateFunctionWithDesc(
	executorName string,
	executorType string,
	colonyName string,
	funcName string,
	description string,
	args []*FunctionArg) *Function {
	return &Function{
		ExecutorName: executorName,
		ExecutorType: executorType,
		ColonyName:   colonyName,
		FuncName:     funcName,
		Description:  description,
		Args:         args,
	}
}

// CreateFunctionArg creates a FunctionArg
func CreateFunctionArg(name string, argType string, description string, required bool, enum []string) *FunctionArg {
	return &FunctionArg{
		Name:        name,
		Type:        argType,
		Description: description,
		Required:    required,
		Enum:        enum,
	}
}

func ConvertJSONToFunction(jsonString string) (*Function, error) {
	var function *Function
	err := json.Unmarshal([]byte(jsonString), &function)
	if err != nil {
		return nil, err
	}

	return function, nil
}

func ConvertJSONToFunctionArray(jsonString string) ([]*Function, error) {
	var functions []*Function
	err := json.Unmarshal([]byte(jsonString), &functions)
	if err != nil {
		return functions, err
	}

	return functions, nil
}

func ConvertFunctionArrayToJSON(functions []*Function) (string, error) {
	jsonBytes, err := json.Marshal(functions)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func IsFunctionArraysEqual(functions1 []*Function, functions2 []*Function) bool {
	counter := 0
	for _, function1 := range functions1 {
		for _, function2 := range functions2 {
			if function1.Equals(function2) {
				counter++
			}
		}
	}

	if counter == len(functions1) && counter == len(functions2) {
		return true
	}

	return false
}

func (function *Function) Equals(function2 *Function) bool {
	if function2 == nil {
		return false
	}

	if function.FunctionID != function2.FunctionID ||
		function.ExecutorName != function2.ExecutorName ||
		function.ExecutorType != function2.ExecutorType ||
		function.ColonyName != function2.ColonyName ||
		function.FuncName != function2.FuncName ||
		function.Description != function2.Description ||
		function.Counter != function2.Counter ||
		function.MinWaitTime != function2.MinWaitTime ||
		function.MaxWaitTime != function2.MaxWaitTime ||
		function.MinExecTime != function2.MinExecTime ||
		function.MaxExecTime != function2.MaxExecTime ||
		function.AvgWaitTime != function2.AvgWaitTime ||
		function.AvgExecTime != function2.AvgExecTime {
		return false
	}

	// Compare Args
	if len(function.Args) != len(function2.Args) {
		return false
	}

	return true
}

func (function *Function) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(function)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
