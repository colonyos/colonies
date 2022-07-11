package core

import (
	"encoding/json"
)

type Conditions struct {
	ColonyID     string   `json:"colonyid"`
	RuntimeIDs   []string `json:"runtimeids"`
	RuntimeType  string   `json:"runtimetype"`
	Mem          int      `json:"mem"`
	Cores        int      `json:"cores"`
	GPUs         int      `json:"gpus"`
	Dependencies []string `json:"dependencies"`
}

type ProcessSpec struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Cmd         string            `json:"cmd"`
	Args        []string          `json:"args"`
	Volumes     []string          `json:"volumes"`
	Ports       []string          `json:"ports"`
	Priority    int               `json:"priority"`
	MaxExecTime int               `json:"maxexectime"`
	MaxRetries  int               `json:"maxretries"`
	Conditions  Conditions        `json:"conditions"`
	Env         map[string]string `json:"env"`
}

func CreateEmptyProcessSpec() *ProcessSpec {
	processSpec := &ProcessSpec{}
	processSpec.Env = make(map[string]string)
	processSpec.Args = make([]string, 0)
	processSpec.Volumes = make([]string, 0)
	processSpec.Ports = make([]string, 0)
	processSpec.MaxExecTime = -1
	processSpec.MaxRetries = -1
	return processSpec
}

func CreateProcessSpec(name string, image string, cmd string, args []string, volumes []string, ports []string, colonyID string, runtimeIDs []string, runtimeType string, maxExecTime int, maxRetries int, mem int, cores int, gpus int, env map[string]string, dependencies []string, priority int) *ProcessSpec {
	conditions := Conditions{ColonyID: colonyID, RuntimeIDs: runtimeIDs, RuntimeType: runtimeType, Mem: mem, Cores: cores, GPUs: gpus, Dependencies: dependencies}
	return &ProcessSpec{Name: name, Image: image, Cmd: cmd, Args: args, Volumes: volumes, Ports: ports, MaxExecTime: maxExecTime, MaxRetries: maxRetries, Conditions: conditions, Env: env, Priority: priority}
}

func ConvertJSONToProcessSpec(jsonString string) (*ProcessSpec, error) {
	processSpec := &ProcessSpec{}
	processSpec.Env = make(map[string]string)

	err := json.Unmarshal([]byte(jsonString), &processSpec)
	if err != nil {
		return nil, err
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
		processSpec.Image != processSpec2.Image ||
		processSpec.Cmd != processSpec2.Cmd ||
		processSpec.MaxExecTime != processSpec2.MaxExecTime ||
		processSpec.MaxRetries != processSpec2.MaxRetries ||
		processSpec.Conditions.ColonyID != processSpec2.Conditions.ColonyID ||
		processSpec.Conditions.RuntimeType != processSpec2.Conditions.RuntimeType ||
		processSpec.Conditions.Mem != processSpec2.Conditions.Mem ||
		processSpec.Conditions.Cores != processSpec2.Conditions.Cores ||
		processSpec.Conditions.GPUs != processSpec2.Conditions.GPUs ||
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

	if processSpec.Volumes != nil && processSpec2.Volumes == nil {
		same = false
	} else if processSpec.Volumes == nil && processSpec2.Volumes != nil {
		same = false
	} else {
		counter := 0
		for _, arg1 := range processSpec.Volumes {
			for _, arg2 := range processSpec2.Volumes {
				if arg1 == arg2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.Volumes) && counter != len(processSpec2.Volumes) {
			same = false
		}
	}

	if processSpec.Ports != nil && processSpec2.Ports == nil {
		same = false
	} else if processSpec.Ports == nil && processSpec2.Ports != nil {
		same = false
	} else {
		counter := 0
		for _, arg1 := range processSpec.Ports {
			for _, arg2 := range processSpec2.Ports {
				if arg1 == arg2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.Ports) && counter != len(processSpec2.Ports) {
			same = false
		}
	}

	if processSpec.Conditions.RuntimeIDs != nil && processSpec2.Conditions.RuntimeIDs == nil {
		same = false
	} else if processSpec.Conditions.RuntimeIDs == nil && processSpec2.Conditions.RuntimeIDs != nil {
		same = false
	} else {
		counter := 0
		for _, targetRuntimeID1 := range processSpec.Conditions.RuntimeIDs {
			for _, targetRuntimeID2 := range processSpec2.Conditions.RuntimeIDs {
				if targetRuntimeID1 == targetRuntimeID2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.Conditions.RuntimeIDs) && counter != len(processSpec2.Conditions.RuntimeIDs) {
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
