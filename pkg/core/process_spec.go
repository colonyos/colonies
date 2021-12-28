package core

import (
	"encoding/json"
)

type Conditions struct {
	ColonyID    string   `json:"colonyid"`
	RuntimeIDs  []string `json:"runtimeids"`
	RuntimeType string   `json:"runtimetype"`
	Mem         int      `json:"mem"`
	Cores       int      `json:"cores"`
	GPUs        int      `json:"gpus"`
}

type ProcessSpec struct {
	Timeout    int               `json:"timeout"`
	MaxRetries int               `json:"maxretries"`
	Conditions Conditions        `json:"conditions"`
	Env        map[string]string `json:"env"`
}

func CreateProcessSpec(colonyID string, runtimeIDs []string, runtimeType string, timeout int, maxRetries int, mem int, cores int, gpus int, env map[string]string) *ProcessSpec {
	conditions := Conditions{ColonyID: colonyID, RuntimeIDs: runtimeIDs, RuntimeType: runtimeType, Mem: mem, Cores: cores, GPUs: gpus}
	return &ProcessSpec{Timeout: timeout, MaxRetries: maxRetries, Conditions: conditions, Env: env}
}

func ConvertJSONToProcessSpec(jsonString string) (*ProcessSpec, error) {
	processSpec := &ProcessSpec{}
	processSpec.Env = make(map[string]string)

	err := json.Unmarshal([]byte(jsonString), &processSpec)
	if err != nil {
		return nil, err
	}

	return processSpec, nil
}

func (processSpec *ProcessSpec) Equals(processSpec2 *ProcessSpec) bool {
	same := true
	if processSpec.Timeout != processSpec2.Timeout &&
		processSpec.MaxRetries != processSpec2.MaxRetries &&
		processSpec.Conditions.ColonyID != processSpec2.Conditions.ColonyID &&
		processSpec.Conditions.RuntimeType != processSpec2.Conditions.RuntimeType &&
		processSpec.Conditions.Mem != processSpec2.Conditions.Mem &&
		processSpec.Conditions.Cores != processSpec2.Conditions.Cores &&
		processSpec.Conditions.GPUs != processSpec2.Conditions.GPUs {
		same = false
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

func (processSpec *ProcessSpec) ToJSON() (string, error) {
	jsonString, err := json.MarshalIndent(processSpec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
