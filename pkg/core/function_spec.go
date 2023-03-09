package core

import (
	"encoding/json"
)

type Conditions struct {
	ColonyID     string   `json:"colonyid"`
	ExecutorIDs  []string `json:"executorids"`
	ExecutorType string   `json:"executortype"`
	Dependencies []string `json:"dependencies"`
}

type FunctionSpec struct {
	NodeName    string            `json:"nodename"`
	FuncName    string            `json:"funcname"`
	Args        []interface{}     `json:"args"`
	Priority    int               `json:"priority"`
	MaxWaitTime int               `json:"maxwaittime"`
	MaxExecTime int               `json:"maxexectime"`
	MaxRetries  int               `json:"maxretries"`
	Conditions  Conditions        `json:"conditions"`
	Label       string            `json:"label"`
	Env         map[string]string `json:"env"`
}

func CreateEmptyFunctionSpec() *FunctionSpec {
	funcSpec := &FunctionSpec{}
	funcSpec.Env = make(map[string]string)
	funcSpec.Args = make([]interface{}, 0)
	funcSpec.MaxExecTime = -1
	funcSpec.MaxRetries = -1
	return funcSpec
}

func CreateFunctionSpec(nodeName string, funcName string, args []interface{}, colonyID string, executorIDs []string, executorType string, maxWaitTime int, maxExecTime int, maxRetries int, env map[string]string, dependencies []string, priority int, label string) *FunctionSpec {
	argsif := make([]interface{}, len(args))
	for k, v := range args {
		argsif[k] = v
	}

	conditions := Conditions{ColonyID: colonyID, ExecutorIDs: executorIDs, ExecutorType: executorType, Dependencies: dependencies}
	return &FunctionSpec{NodeName: nodeName, FuncName: funcName, Args: argsif, MaxWaitTime: maxWaitTime, MaxExecTime: maxExecTime, MaxRetries: maxRetries, Conditions: conditions, Env: env, Priority: priority, Label: label}
}

func ConvertJSONToFunctionSpec(jsonString string) (*FunctionSpec, error) {
	funcSpec := &FunctionSpec{}
	funcSpec.Env = make(map[string]string)

	err := json.Unmarshal([]byte(jsonString), &funcSpec)
	if err != nil {
		return nil, err
	}

	if funcSpec.MaxWaitTime == 0 {
		funcSpec.MaxWaitTime = -1
	}

	if funcSpec.MaxExecTime == 0 {
		funcSpec.MaxExecTime = -1
	}

	return funcSpec, nil
}

func (funcSpec *FunctionSpec) Equals(funcSpec2 *FunctionSpec) bool {
	if funcSpec2 == nil {
		return false
	}

	same := true
	if funcSpec.NodeName != funcSpec2.NodeName ||
		funcSpec.FuncName != funcSpec2.FuncName ||
		funcSpec.MaxWaitTime != funcSpec2.MaxWaitTime ||
		funcSpec.MaxExecTime != funcSpec2.MaxExecTime ||
		funcSpec.MaxRetries != funcSpec2.MaxRetries ||
		funcSpec.Conditions.ColonyID != funcSpec2.Conditions.ColonyID ||
		funcSpec.Conditions.ExecutorType != funcSpec2.Conditions.ExecutorType ||
		funcSpec.Priority != funcSpec2.Priority ||
		funcSpec.Label != funcSpec2.Label {
		same = false
	}

	if funcSpec.Args != nil && funcSpec2.Args == nil {
		same = false
	} else if funcSpec.Args == nil && funcSpec2.Args != nil {
		same = false
	} else {
		counter := 0
		for _, arg1 := range funcSpec.Args {
			for _, arg2 := range funcSpec2.Args {
				if arg1 == arg2 {
					counter++
				}
			}
		}
		if counter != len(funcSpec.Args) && counter != len(funcSpec2.Args) {
			same = false
		}
	}

	if funcSpec.Conditions.ExecutorIDs != nil && funcSpec2.Conditions.ExecutorIDs == nil {
		same = false
	} else if funcSpec.Conditions.ExecutorIDs == nil && funcSpec2.Conditions.ExecutorIDs != nil {
		same = false
	} else {
		counter := 0
		for _, targetExecutorID1 := range funcSpec.Conditions.ExecutorIDs {
			for _, targetExecutorID2 := range funcSpec2.Conditions.ExecutorIDs {
				if targetExecutorID1 == targetExecutorID2 {
					counter++
				}
			}
		}
		if counter != len(funcSpec.Conditions.ExecutorIDs) && counter != len(funcSpec2.Conditions.ExecutorIDs) {
			same = false
		}
	}

	if funcSpec.Conditions.Dependencies != nil && funcSpec2.Conditions.Dependencies == nil {
		same = false
	} else if funcSpec.Conditions.Dependencies == nil && funcSpec2.Conditions.Dependencies != nil {
		same = false
	} else {
		counter := 0
		for _, dependency1 := range funcSpec.Conditions.Dependencies {
			for _, dependency2 := range funcSpec2.Conditions.Dependencies {
				if dependency1 == dependency2 {
					counter++
				}
			}
		}
		if counter != len(funcSpec.Conditions.Dependencies) && counter != len(funcSpec2.Conditions.Dependencies) {
			same = false
		}
	}

	if funcSpec.Env != nil && funcSpec2.Env == nil {
		same = false
	} else if funcSpec.Env == nil && funcSpec2.Env != nil {
		same = false
	} else {
		counter := 0
		for k, v := range funcSpec.Env {
			if funcSpec2.Env[k] == v {
				counter++
			}
		}

		if !(counter == len(funcSpec.Env) && counter == len(funcSpec2.Env)) {
			same = false
		}
	}

	return same
}

func (funcSpec *FunctionSpec) AddDependency(dependency string) {
	funcSpec.Conditions.Dependencies = append(funcSpec.Conditions.Dependencies, dependency)
}

func (funcSpec *FunctionSpec) ToJSON() (string, error) {
	jsonString, err := json.MarshalIndent(funcSpec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
