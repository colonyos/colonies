package core

import (
	"encoding/json"
)

type Conditions struct {
	RuntimeType string `json:"runtimetype"`
	Mem         int    `json:"mem"`
	Cores       int    `json:"cores"`
	GPUs        int    `json:"gpus"`
}

type ProcessSpec struct {
	TargetColonyID   string            `json:"targetcolonyid"`
	TargetRuntimeIDs []string          `json:"targetruntimeids"`
	Timeout          int               `json:"timeout"`
	MaxRetries       int               `json:"maxretries"`
	Conditions       Conditions        `json:"conditions"`
	Env              map[string]string `json:"env"`
}

func CreateProcessSpec(targetColonyID string, targetRuntimeIDs []string, runtimeType string, timeout int, maxRetries int, mem int, cores int, gpus int, env map[string]string) *ProcessSpec {
	conditions := Conditions{RuntimeType: runtimeType, Mem: mem, Cores: cores, GPUs: gpus}
	return &ProcessSpec{TargetColonyID: targetColonyID, TargetRuntimeIDs: targetRuntimeIDs, Timeout: timeout, Conditions: conditions, Env: env}
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
	if processSpec.TargetColonyID != processSpec2.TargetColonyID &&
		processSpec.Timeout != processSpec2.Timeout &&
		processSpec.MaxRetries != processSpec2.MaxRetries &&
		processSpec.Conditions.RuntimeType != processSpec2.Conditions.RuntimeType &&
		processSpec.Conditions.Mem != processSpec2.Conditions.Mem &&
		processSpec.Conditions.Cores != processSpec2.Conditions.Cores &&
		processSpec.Conditions.GPUs != processSpec2.Conditions.GPUs {
		same = false
	}

	if processSpec.TargetRuntimeIDs != nil && processSpec2.TargetRuntimeIDs == nil {
		same = false
	} else if processSpec.TargetRuntimeIDs == nil && processSpec2.TargetRuntimeIDs != nil {
		same = false
	} else {
		counter := 0
		for _, targetRuntimeID1 := range processSpec.TargetRuntimeIDs {
			for _, targetRuntimeID2 := range processSpec2.TargetRuntimeIDs {
				if targetRuntimeID1 == targetRuntimeID2 {
					counter++
				}
			}
		}
		if counter != len(processSpec.TargetRuntimeIDs) && counter != len(processSpec2.TargetRuntimeIDs) {
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
