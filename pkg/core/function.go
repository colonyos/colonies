package core

import "encoding/json"

type Function struct {
	ID   string   `json:"funcid"`
	Name string   `json:"funcname"`
	Desc string   `json:"funcdesc"`
	Args []string `json:"args"`
}

func CreateFunction(id string,
	name string,
	desc string,
	args []string) Function {
	return Function{ID: id,
		Name: name,
		Desc: desc,
		Args: args,
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

	if function.ID != function2.ID || function.Name != function2.Name || function.Desc != function2.Desc {
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

	return false
}

func (function *Function) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(function, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
