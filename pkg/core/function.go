package core

import "encoding/json"

type Function struct {
	FunctionID  string   `json:"functionid"`
	ExecutorID  string   `json:"executorid"`
	ColonyID    string   `json:"colonyid"`
	Name        string   `json:"name"`
	Desc        string   `json:"desc"`
	AvgWaitTime float64  `json:"avgwaittime"`
	AvgExecTime float64  `json:"avgexectime"`
	Args        []string `json:"args"`
}

func CreateFunction(functionID string,
	executorID string,
	colonyID string,
	name string,
	desc string,
	avgWaitTime float64,
	avgExecTime float64,
	args []string) Function {
	return Function{
		FunctionID:  functionID,
		ExecutorID:  executorID,
		ColonyID:    colonyID,
		Name:        name,
		Desc:        desc,
		AvgWaitTime: avgWaitTime,
		AvgExecTime: avgExecTime,
		Args:        args,
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

func ConvertJSONToFunctionArray(jsonString string) ([]Function, error) {
	var functions []Function
	err := json.Unmarshal([]byte(jsonString), &functions)
	if err != nil {
		return functions, err
	}

	return functions, nil
}

func ConvertFunctionArrayToJSON(functions []Function) (string, error) {
	jsonBytes, err := json.MarshalIndent(functions, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func IsFunctionArraysEqual(functions1 []Function, functions2 []Function) bool {
	counter := 0
	for _, function1 := range functions1 {
		for _, function2 := range functions2 {
			if function1.Equals(&function2) {
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
		function.ExecutorID != function2.ExecutorID ||
		function.ColonyID != function2.ColonyID ||
		function.Name != function2.Name ||
		function.Desc != function2.Desc ||
		function.AvgWaitTime != function2.AvgWaitTime ||
		function.AvgExecTime != function2.AvgExecTime {
		return false
	}

	counter := 0
	for _, arg1 := range function.Args {
		for _, arg2 := range function2.Args {
			if arg1 == arg2 {
				counter++
			}
		}
	}
	if counter != len(function.Args) || counter != len(function2.Args) {
		return false
	}

	return true
}

func (function *Function) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(function, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
