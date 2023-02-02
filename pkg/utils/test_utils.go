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
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, "test_executor_type", 1000, 100, 1, make(map[string]string), []string{}, 1)
}

func CreateTestProcessWithType(colonyID string, executorType string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithType(colonyID, executorType))
}

func CreateTestProcessSpecWithType(colonyID string, executorType string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, executorType, 1000, 100, 1, make(map[string]string), []string{}, 1)
}

func CreateTestProcessWithEnv(colonyID string, env map[string]string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithEnv(colonyID, env))
}

func CreateTestProcessSpecWithEnv(colonyID string, env map[string]string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, []string{}, "test_executor_type", 1000, 100, 1, env, []string{}, 1)
}

func CreateTestProcessWithTargets(colonyID string, targetExecutorIDs []string) *core.Process {
	return core.CreateProcess(CreateTestProcessSpecWithTargets(colonyID, targetExecutorIDs))
}

func CreateTestProcessSpecWithTargets(colonyID string, targetExecutorIDs []string) *core.ProcessSpec {
	return core.CreateProcessSpec("test_name", "test_func", []string{"test_arg"}, colonyID, targetExecutorIDs, "test_executor_type", 1000, 100, 1, make(map[string]string), []string{}, 1)
}

func CreateTestExecutor(colonyID string) *core.Executor {
	return core.CreateExecutor(core.GenerateRandomID(), "test_executor_type", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestExecutorWithType(colonyID string, executorType string) *core.Executor {
	return core.CreateExecutor(core.GenerateRandomID(), executorType, core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestExecutorWithID(colonyID string, executorID string) *core.Executor {
	return core.CreateExecutor(executorID, "test_executor_type", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())
}

func CreateTestExecutorWithKey(colonyID string) (*core.Executor, string, error) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	executorID, err := crypto.GenerateID(executorPrvKey)
	if err != nil {
		return nil, "", err
	}

	return core.CreateExecutor(executorID, "test_executor_type", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now()), executorPrvKey, nil
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
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10)
	return generator
}

func FakeGeneratorSingleProcess(t *testing.T, colonyID string) *core.Generator {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec1 := CreateTestProcessSpec(colonyID)
	processSpec1.Name = "task1"
	workflowSpec.AddProcessSpec(processSpec1)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10)
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
	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	return cron
}

func FakeSingleCron(t *testing.T, colonyID string) *core.Cron {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec := CreateTestProcessSpec(colonyID)
	processSpec.Name = "task1"
	workflowSpec.AddProcessSpec(processSpec)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	return cron
}
