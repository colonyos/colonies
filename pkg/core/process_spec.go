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

type ProcessSpec struct {
	Name        string            `json:"name"`
	Func        string            `json:"func"`
	Args        []string          `json:"args"`
	Priority    int               `json:"priority"`
	MaxWaitTime int               `json:"maxwaittime"`
	MaxExecTime int               `json:"maxexectime"`
	MaxRetries  int               `json:"maxretries"`
	Conditions  Conditions        `json:"conditions"`
	Env         map[string]string `json:"env"`
}

func CreateEmptyProcessSpec() *ProcessSpec {
	processSpec := &ProcessSpec{}
	processSpec.Env = make(map[string]string)
	processSpec.Args = make([]string, 0)
	processSpec.MaxExecTime = -1
	processSpec.MaxRetries = -1
	return processSpec
}

func CreateProcessSpec(name string, fn string, args []string, colonyID string, executorIDs []string, executorType string, maxWaitTime int, maxExecTime int, maxRetries int, env map[string]string, dependencies []string, priority int) *ProcessSpec {
	conditions := Conditions{ColonyID: colonyID, ExecutorIDs: executorIDs, ExecutorType: executorType, Dependencies: dependencies}
	return &ProcessSpec{Name: name, Func: fn, Args: args, MaxWaitTime: maxWaitTime, MaxExecTime: maxExecTime, MaxRetries: maxRetries, Conditions: conditions, Env: env, Priority: priority}
}

func ConvertJSONToProcessSpec(jsonString string) (*ProcessSpec, error) {
	processSpec := &ProcessSpec{}
	processSpec.Env = make(map[string]string)

	err := json.Unmarshal([]byte(jsonString), &processSpec)
	if err != nil {
		return nil, err
	}

	if processSpec.MaxWaitTime == 0 {
		processSpec.MaxWaitTime = -1
	}

	if processSpec.MaxExecTime == 0 {
		processSpec.MaxExecTime = -1
	}

	return processSpec, nil
}

func (processSpec *ProcessSpec) Equals(processSpec2 *ProcessSpec) bool {
	if processSpec2 == nil {
		return false
	}

	same := true
	if processSpec.Name != processSpec2.Name ||
		processSpec.Func != processSpec2.Func ||
		processSpec.MaxWaitTime != processSpec2.MaxWaitTime ||
		processSpec.MaxExecTime != processSpec2.MaxExecTime ||
		processSpec.MaxRetries != processSpec2.MaxRetries ||
		processSpec.Conditions.ColonyID != processSpec2.Conditions.ColonyID ||
		processSpec.Conditions.ExecutorType != processSpec2.Conditions.ExecutorType ||
		processSpec.Priority != processSpec2.Priority {
		same = false
	}

	if processSpec.Args != nil && processSpec2.Args == nil {
		same = false
	} else if processSpec.Args == nil && processSpec2.Args != nil {
		same = false
	} else {
		counter := 0
		for _, arg1 := range processSpec.Args {
			for _, arg2 := range processSpec2.Args {
				if arg1 == arg2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.Args) && counter != len(processSpec2.Args) {
			same = false
		}
	}

	if processSpec.Conditions.ExecutorIDs != nil && processSpec2.Conditions.ExecutorIDs == nil {
		same = false
	} else if processSpec.Conditions.ExecutorIDs == nil && processSpec2.Conditions.ExecutorIDs != nil {
		same = false
	} else {
		counter := 0
		for _, targetExecutorID1 := range processSpec.Conditions.ExecutorIDs {
			for _, targetExecutorID2 := range processSpec2.Conditions.ExecutorIDs {
				if targetExecutorID1 == targetExecutorID2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.Conditions.ExecutorIDs) && counter != len(processSpec2.Conditions.ExecutorIDs) {
			same = false
		}
	}

	if processSpec.Conditions.Dependencies != nil && processSpec2.Conditions.Dependencies == nil {
		same = false
	} else if processSpec.Conditions.Dependencies == nil && processSpec2.Conditions.Dependencies != nil {
		same = false
	} else {
		counter := 0
		for _, dependency1 := range processSpec.Conditions.Dependencies {
			for _, dependency2 := range processSpec2.Conditions.Dependencies {
				if dependency1 == dependency2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.Conditions.Dependencies) && counter != len(processSpec2.Conditions.Dependencies) {
			same = false
		}
	}

	if processSpec.Env != nil && processSpec2.Env == nil {
		same = false
	} else if processSpec.Env == nil && processSpec2.Env != nil {
		same = false
	} else {
		counter := 0
		for k, v := range processSpec.Env {
			if processSpec2.Env[k] == v {
				counter++
			}
		}

		if !(counter == len(processSpec.Env) && counter == len(processSpec2.Env)) {
			same = false
		}
	}

	return same
}

func (processSpec *ProcessSpec) AddDependency(dependency string) {
	processSpec.Conditions.Dependencies = append(processSpec.Conditions.Dependencies, dependency)
}

func (processSpec *ProcessSpec) ToJSON() (string, error) {
	jsonString, err := json.MarshalIndent(processSpec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
