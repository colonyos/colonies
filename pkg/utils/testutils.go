package utils

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func CreateTestProcess(colonyID string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpec(colonyID))
}

func CreateTestProcessSpec(colonyID string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, "test_runtime_type", 1000, 100, 1, make(map[string]string), []string{}, 1)
}

func CreateTestProcessWithType(colonyID string, runtimeType string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithType(colonyID, runtimeType))
}

func CreateTestProcessSpecWithType(colonyID string, runtimeType string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, runtimeType, 1000, 100, 1, make(map[string]string), []string{}, 1)
}

func CreateTestProcessWithEnv(colonyID string, env map[string]string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithEnv(colonyID, env))
}

func CreateTestProcessSpecWithEnv(colonyID string, env map[string]string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, "test_runtime_type", 1000, 100, 1, env, []string{}, 1)
}

func CreateTestProcessWithTargets(colonyID string, targetRuntimeIDs []string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithTargets(colonyID, targetRuntimeIDs))
}

func CreateTestProcessSpecWithTargets(colonyID string, targetRuntimeIDs []string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, targetRuntimeIDs, "test_runtime_type", 1000, 100, 1, make(map[string]string), []string{}, 1)
}

func CreateTestRuntime(colonyID string) *core.Runtime {
	return core.CreateRuntime(core.GenerateRandomID(), "test_runtime_type", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestRuntimeWithType(colonyID string, runtimeType string) *core.Runtime {
	return core.CreateRuntime(core.GenerateRandomID(), runtimeType, core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestRuntimeWithID(colonyID string, runtimeID string) *core.Runtime {
	return core.CreateRuntime(runtimeID, "test_runtime_type", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestRuntimeWithKey(colonyID string) (*core.Runtime, string, error) {
	crypto := crypto.CreateCrypto()
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	if err != nil {
		return nil, "", err
	}

	return core.CreateRuntime(runtimeID, "test_runtime_type", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now()), runtimePrvKey, nil
}

func CreateTestColonyWithKey() (*core.Colony, string, error) {
	crypto := crypto.CreateCrypto()

	colonyPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	colonyID, err := crypto.GenerateID(colonyPrvKey)
	if err != nil {
		return nil, "", err
	}
	return core.CreateColony(colonyID, "test_colony_name"), colonyPrvKey, nil
}

func FakeGenerator(t *testing.T, colonyID string) *core.Generator {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec1 := CreateTestProcessSpec(colonyID)
	processSpec1.Name = "task1"
	processSpec2 := CreateTestProcessSpec(colonyID)
	processSpec2.Name = "task2"
	processSpec2.AddDependency("task1")
	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := core.CreateGenerator(colonyID, "test_genname", jsonStr, 10, 0, 1)
	return generator
}

func FakeCron(t *testing.T, colonyID string) *core.Cron {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec1 := CreateTestProcessSpec(colonyID)
	processSpec1.Name = "task1"
	processSpec2 := CreateTestProcessSpec(colonyID)
	processSpec2.Name = "task2"
	processSpec2.AddDependency("task1")
	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	cron := core.CreateCron(colonyID, "test_cron1", "1 * * * * *", -1, false, jsonStr)
	return cron
}
