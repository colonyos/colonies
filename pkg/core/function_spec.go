package core

import (
	"encoding/json"
)

type Filesystem struct {
	Mount          string          `json:"mount"`
	SnapshotMounts []SnapshotMount `json:"snapshots"`
	SyncDirMounts  []SyncDirMount  `json:"dirs"`
}

type SnapshotMount struct {
	SnapshotID  string `json:"snapshotid"`
	Label       string `json:"label"`
	Dir         string `json:"dir"`
	KeepFiles   bool   `json:"keepfiles"`
	KeepSnaphot bool   `json:"keepsnapshot"`
}

type OnStart struct {
	KeepLocal bool `json:"keeplocal"`
}

type OnClose struct {
	KeepLocal bool `json:"keeplocal"`
}

type ConflictResolution struct {
	OnStart OnStart `json:"onstart"`
	OnClose OnClose `json:"onclose"`
}

type SyncDirMount struct {
	Label              string             `json:"label"`
	Dir                string             `json:"dir"`
	KeepFiles          bool               `json:"keepfiles"`
	ConflictResolution ConflictResolution `json:"onconflicts"`
}

type Conditions struct {
	ColonyName       string   `json:"colonyname"`
	ExecutorNames    []string `json:"executornames"`
	ExecutorType     string   `json:"executortype"`
	LocationName     string   `json:"locationname,omitempty"` // Optional filter by location
	Dependencies     []string `json:"dependencies"`
	Nodes            int      `json:"nodes"`
	CPU              string   `json:"cpu"`
	Processes        int      `json:"processes"`
	ProcessesPerNode int      `json:"processespernode"`
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
	Filesystem  Filesystem             `json:"fs"`
	Env      map[string]string `json:"env"`
	Channels []string          `json:"channels,omitempty"`
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

func CreateFunctionSpec(nodeName string, funcName string, args []interface{}, kwargs map[string]interface{}, colonyName string, executorNames []string, executorType string, maxWaitTime int, maxExecTime int, maxRetries int, env map[string]string, dependencies []string, priority int, label string) *FunctionSpec {
	argsif := make([]interface{}, len(args))
	for k, v := range args {
		argsif[k] = v
	}

	kwargsif := make(map[string]interface{}, len(kwargs))
	for k, v := range kwargs {
		kwargsif[k] = v
	}

	conditions := Conditions{ColonyName: colonyName, ExecutorNames: executorNames, ExecutorType: executorType, Dependencies: dependencies}
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
		funcSpec.Conditions.ColonyName != funcSpec2.Conditions.ColonyName ||
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

	if funcSpec.Conditions.ExecutorNames != nil && funcSpec2.Conditions.ExecutorNames == nil {
		same = false
	} else if funcSpec.Conditions.ExecutorNames == nil && funcSpec2.Conditions.ExecutorNames != nil {
		same = false
	} else {
		counter := 0
		for _, targetExecutorName1 := range funcSpec.Conditions.ExecutorNames {
			for _, targetExecutorName2 := range funcSpec2.Conditions.ExecutorNames {
				if targetExecutorName1 == targetExecutorName2 {
					counter++
				}
			}
		}
		if counter != len(funcSpec.Conditions.ExecutorNames) && counter != len(funcSpec2.Conditions.ExecutorNames) {
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

	if funcSpec.Filesystem.Mount != funcSpec2.Filesystem.Mount {
		same = false
	}

	if len(funcSpec.Filesystem.SyncDirMounts) != len(funcSpec2.Filesystem.SyncDirMounts) {
		return false
	}

	if len(funcSpec.Filesystem.SnapshotMounts) != len(funcSpec2.Filesystem.SnapshotMounts) {
		return false
	}

	for i := range funcSpec.Filesystem.SyncDirMounts {
		if funcSpec.Filesystem.SyncDirMounts[i].Label != funcSpec2.Filesystem.SyncDirMounts[i].Label {
			same = false
		}
		if funcSpec.Filesystem.SyncDirMounts[i].Dir != funcSpec2.Filesystem.SyncDirMounts[i].Dir {
			same = false
		}
		if funcSpec.Filesystem.SyncDirMounts[i].KeepFiles != funcSpec2.Filesystem.SyncDirMounts[i].KeepFiles {
			same = false
		}
	}

	for i := range funcSpec.Filesystem.SnapshotMounts {
		if funcSpec.Filesystem.SnapshotMounts[i].SnapshotID != funcSpec2.Filesystem.SnapshotMounts[i].SnapshotID {
			same = false
		}
		if funcSpec.Filesystem.SnapshotMounts[i].Dir != funcSpec2.Filesystem.SnapshotMounts[i].Dir {
			same = false
		}
		if funcSpec.Filesystem.SnapshotMounts[i].Label != funcSpec2.Filesystem.SnapshotMounts[i].Label {
			same = false
		}
		if funcSpec.Filesystem.SnapshotMounts[i].KeepFiles != funcSpec2.Filesystem.SnapshotMounts[i].KeepFiles {
			same = false
		}
		if funcSpec.Filesystem.SnapshotMounts[i].KeepSnaphot != funcSpec2.Filesystem.SnapshotMounts[i].KeepSnaphot {
			same = false
		}
	}

	return same
}

func (funcSpec *FunctionSpec) AddDependency(dependency string) {
	funcSpec.Conditions.Dependencies = append(funcSpec.Conditions.Dependencies, dependency)
}

func (funcSpec *FunctionSpec) ToJSON() (string, error) {
	jsonString, err := json.Marshal(funcSpec)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
