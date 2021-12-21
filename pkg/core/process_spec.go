package core

import "encoding/json"

type ProcessSpec struct {
	TargetColonyID    string            `json:"targetcolonyid"`
	TargetComputerIDs []string          `json:"targetcomputerids"`
	ComputerType      string            `json:"computertype"`
	Timeout           int               `json:"timeout"`
	MaxRetries        int               `json:"maxretries"`
	Mem               int               `json:"mem"`
	Cores             int               `json:"cores"`
	GPUs              int               `json:"gpus"`
	In                map[string]string `json:"in"`
}

func CreateProcessSpec(targetColonyID string, targetComputerIDs []string, computerType string, timeout int, maxRetries int, mem int, cores int, gpus int, in map[string]string) *ProcessSpec {
	return &ProcessSpec{TargetColonyID: targetColonyID, TargetComputerIDs: targetComputerIDs, ComputerType: computerType, Timeout: timeout, Mem: mem, Cores: cores, GPUs: gpus, In: in}
}

func ConvertJSONToProcessSpec(jsonString string) (*ProcessSpec, error) {
	processSpec := &ProcessSpec{}
	processSpec.In = make(map[string]string)

	err := json.Unmarshal([]byte(jsonString), &processSpec)
	if err != nil {
		return nil, err
	}

	return processSpec, nil
}

func (processSpec *ProcessSpec) ToJSON() (string, error) {
	jsonString, err := json.MarshalIndent(processSpec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
