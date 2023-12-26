package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEmptyFunctionSpec(t *testing.T) {
	funcSpec := CreateEmptyFunctionSpec()
	assert.NotNil(t, funcSpec)
}

func TestFunctionSpecJSON(t *testing.T) {
	colonyName := GenerateRandomID()
	executorType := "test_executor_type"
	executor1Name := GenerateRandomID()
	executor2Name := GenerateRandomID()
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3
	env := make(map[string]string)
	env["test_key"] = "test_value"

	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["0"] = "test_arg"

	var snapshots []SnapshotMount
	snapshot1 := SnapshotMount{Label: "test_label1", SnapshotID: "test_snapshotid1", Dir: "test_dir1", KeepFiles: false, KeepSnaphot: false}
	snapshot2 := SnapshotMount{Label: "test_label2", SnapshotID: "test_snapshotid2", Dir: "test_dir2", KeepFiles: true, KeepSnaphot: true}
	snapshots = append(snapshots, snapshot1)
	snapshots = append(snapshots, snapshot2)
	var syncdirs []SyncDirMount
	syncdir1 := SyncDirMount{Label: "test_label1", Dir: "test_dir1", KeepFiles: false}
	syncdir2 := SyncDirMount{Label: "test_label2", Dir: "test_dir2", KeepFiles: false}
	syncdirs = append(syncdirs, syncdir1)
	syncdirs = append(syncdirs, syncdir2)

	funcSpec := CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyName, []string{executor1Name, executor2Name}, executorType, maxWaitTime, maxExecTime, maxRetries, env, []string{"test_name2"}, 5, "test_label")
	funcSpec.Filesystem = Filesystem{SnapshotMounts: snapshots, SyncDirMounts: syncdirs, Mount: "/cfs"}

	funcSpec.Conditions.Nodes = 10
	funcSpec.Conditions.CPU = "1000m"
	funcSpec.Conditions.Processes = 10
	funcSpec.Conditions.ProcessesPerNode = 1
	funcSpec.Conditions.Memory = "10G"
	funcSpec.Conditions.Storage = "10G"
	funcSpec.Conditions.GPU = GPU{Name: "test_name1", Count: 1, Memory: "11G"}
	funcSpec.Conditions.WallTime = 1000

	jsonString, err := funcSpec.ToJSON()
	assert.Nil(t, err)

	funcSpec2, err := ConvertJSONToFunctionSpec(jsonString + "error")
	assert.NotNil(t, err)

	funcSpec2, err = ConvertJSONToFunctionSpec(jsonString)
	assert.Nil(t, err)

	assert.Equal(t, funcSpec.Conditions.ColonyName, funcSpec2.Conditions.ColonyName)
	assert.Equal(t, funcSpec.MaxExecTime, funcSpec2.MaxExecTime)
	assert.Equal(t, funcSpec.MaxRetries, funcSpec2.MaxRetries)
	assert.Equal(t, funcSpec.Conditions.ExecutorNames, funcSpec2.Conditions.ExecutorNames)
	assert.Contains(t, funcSpec.Conditions.ExecutorNames, executor1Name)
	assert.Contains(t, funcSpec.Conditions.ExecutorNames, executor2Name)
	assert.Equal(t, funcSpec.Conditions.ExecutorType, funcSpec2.Conditions.ExecutorType)
	assert.Equal(t, funcSpec.Env, funcSpec2.Env)
}

func TestFunctionSpecEquals(t *testing.T) {
	colonyName := GenerateRandomID()
	executorType := "test_executor_type"
	executor1Name := GenerateRandomID()
	executor2Name := GenerateRandomID()
	executor3Name := GenerateRandomID()
	maxWaitTime := -1
	maxExecTime := -1
	maxRetries := 3
	env := make(map[string]string)
	env["test_key"] = "test_value"

	env2 := make(map[string]string)
	env2["test_key2"] = "test_value2"

	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["0"] = "test_arg"

	var snapshots []SnapshotMount
	snapshot1 := SnapshotMount{Label: "test_label1", SnapshotID: "test_snapshotid1", Dir: "test_dir1", KeepFiles: false, KeepSnaphot: false}
	snapshot2 := SnapshotMount{Label: "test_label2", SnapshotID: "test_snapshotid2", Dir: "test_dir2", KeepFiles: true, KeepSnaphot: true}
	snapshots = append(snapshots, snapshot1)
	snapshots = append(snapshots, snapshot2)
	var syncdirs []SyncDirMount
	syncdir1 := SyncDirMount{Label: "test_label1", Dir: "test_dir1", KeepFiles: false}
	syncdir2 := SyncDirMount{Label: "test_label2", Dir: "test_dir2", KeepFiles: false}
	syncdirs = append(syncdirs, syncdir1)
	syncdirs = append(syncdirs, syncdir2)

	funcSpec1 := CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyName, []string{executor1Name, executor2Name}, executorType, maxWaitTime, maxExecTime, maxRetries, env, []string{}, 1, "test_label")
	funcSpec1.Filesystem = Filesystem{SnapshotMounts: snapshots, SyncDirMounts: syncdirs, Mount: "/cfs"}

	args = make([]interface{}, 1)
	args[0] = "test_arg2"

	functionSpec2 := CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyName, []string{executor3Name}, executorType+"2", 200, 4, 2, env2, []string{}, 1, "test_label")

	assert.True(t, funcSpec1.Equals(funcSpec1))
	assert.False(t, funcSpec1.Equals(nil))
	assert.False(t, funcSpec1.Equals(functionSpec2))
}
