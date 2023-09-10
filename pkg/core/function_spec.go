package core

import (
	"encoding/json"
)

type SyncDir struct {
	Label            string `json:"label"`
	SnapshotID       string `json:"snapshotid"`
	Dir              string `json:"dir"`
	SyncOnCompletion bool   `json:"sync_on_completion"`
}

type Conditions struct {
	ColonyID         string   `json:"colonyid"`
	ExecutorIDs      []string `json:"executorids"`
	ExecutorType     string   `json:"executortype"`
	Dependencies     []string `json:"dependencies"`
	Nodes            int      `json:"nodes"`
	CPU              string   `json:"cpu"`
	Processes        int      `json:"processes"`
	ProcessesPerNode int      `json:"processes_per_node"`
	Memory           string   `json:"mem"`
	Storage          string   `json:"storage"`
	GPU              GPU      `json:"gpu"`
	WallTime         int64    `json:"walltime"`
}

type FunctionSpec struct {
	NodeName    string                 `json:"nodename"`
	FuncName    string                 `json:"funcname"`
	Args        []interface{}          `json:"args"`
	KwArgs      map[string]interface{} `json:"kwargs"`
	Priority    int                    `json:"priority"`
	MaxWaitTime int                    `json:"maxwaittime"`
	MaxExecTime int                    `json:"maxexectime"`
	MaxRetries  int                    `json:"maxretries"`
	Conditions  Conditions             `json:"conditions"`
	Label       string                 `json:"label"`
	Filesystem  []*SyncDir             `json:"fs"`
	Env         map[string]string      `json:"env"`
}

func CreateEmptyFunctionSpec() *FunctionSpec {
	funcSpec := &FunctionSpec{}
	funcSpec.Env = make(map[string]string)
	funcSpec.Args = make([]interface{}, 0)
	funcSpec.KwArgs = make(map[string]interface{}, 0)
	funcSpec.MaxExecTime = -1
	funcSpec.MaxRetries = -1
	return funcSpec
}

func CreateFunctionSpec(nodeName string, funcName string, args []interface{}, kwargs map[string]interface{}, colonyID string, executorIDs []string, executorType string, maxWaitTime int, maxExecTime int, maxRetries int, env map[string]string, dependencies []string, priority int, label string) *FunctionSpec {
	argsif := make([]interface{}, len(args))
	for k, v := range args {
		argsif[k] = v
	}

	kwargsif := make(map[string]interface{}, len(kwargs))
	for k, v := range kwargs {
		kwargsif[k] = v
	}

	conditions := Conditions{ColonyID: colonyID, ExecutorIDs: executorIDs, ExecutorType: executorType, Dependencies: dependencies}
	return &FunctionSpec{NodeName: nodeName, FuncName: funcName, Args: argsif, KwArgs: kwargsif, MaxWaitTime: maxWaitTime, MaxExecTime: maxExecTime, MaxRetries: maxRetries, Conditions: conditions, Env: env, Priority: priority, Label: label}
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

	if funcSpec.KwArgs != nil && funcSpec2.KwArgs == nil {
		same = false
	} else if funcSpec.KwArgs == nil && funcSpec2.KwArgs != nil {
		same = false
	} else {
		counter := 0
		for name1, arg1 := range funcSpec.KwArgs {
			for name2, arg2 := range funcSpec2.KwArgs {
				if arg1 == arg2 && name1 == name2 {
					counter++
				}
			}
		}
		if counter != len(funcSpec.KwArgs) && counter != len(funcSpec2.KwArgs) {
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
		if funcSpec.Conditions.GPU.Name != funcSpec2.Conditions.GPU.Name {
			same = false
		}
		if funcSpec.Conditions.GPU.Count != funcSpec2.Conditions.GPU.Count {
			same = false
		}
		if funcSpec.Conditions.GPU.Memory != funcSpec2.Conditions.GPU.Memory {
			same = false
		}
		if funcSpec.Conditions.CPU != funcSpec2.Conditions.CPU {
			same = false
		}
		if funcSpec.Conditions.Processes != funcSpec2.Conditions.Processes {
			same = false
		}
		if funcSpec.Conditions.ProcessesPerNode != funcSpec2.Conditions.ProcessesPerNode {
			same = false
		}
		if funcSpec.Conditions.Memory != funcSpec2.Conditions.Memory {
			same = false
		}
		if funcSpec.Conditions.Storage != funcSpec2.Conditions.Storage {
			same = false
		}
		if funcSpec.Conditions.Nodes != funcSpec2.Conditions.Nodes {
			same = false
		}
		if funcSpec.Conditions.WallTime != funcSpec2.Conditions.WallTime {
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

	if len(funcSpec.Filesystem) != len(funcSpec2.Filesystem) {
		return false
	}

	for i := range funcSpec.Filesystem {
		if funcSpec.Filesystem[i].Label != funcSpec2.Filesystem[i].Label {
			same = false
		}
		if funcSpec.Filesystem[i].SnapshotID != funcSpec2.Filesystem[i].SnapshotID {
			same = false
		}
		if funcSpec.Filesystem[i].Dir != funcSpec2.Filesystem[i].Dir {
			same = false
		}
		if funcSpec.Filesystem[i].SyncOnCompletion != funcSpec2.Filesystem[i].SyncOnCompletion {
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
