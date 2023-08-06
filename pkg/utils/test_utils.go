package utils

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func CreateTestProcess(colonyID string) *core.Process {
	return core.CreateProcess(CreateTestFunctionSpec(colonyID))
}

func CreateTestFunctionSpec(colonyID string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, []string{}, "test_executor_type", 1000, 100, 1, make(map[string]string), []string{}, 1, "test_label")
}

func CreateTestProcessWithType(colonyID string, executorType string) *core.Process {
	return core.CreateProcess(CreateTestFunctionSpecWithType(colonyID, executorType))
}

func CreateTestFunctionSpecWithType(colonyID string, executorType string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, []string{}, executorType, 1000, 100, 1, make(map[string]string), []string{}, 1, "test_label")
}

func CreateTestProcessWithEnv(colonyID string, env map[string]string) *core.Process {
	return core.CreateProcess(CreateTestFunctionSpecWithEnv(colonyID, env))
}

func CreateTestFunctionSpecWithEnv(colonyID string, env map[string]string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, []string{}, "test_executor_type", 1000, 100, 1, env, []string{}, 1, "test_label")
}

func CreateTestProcessWithTargets(colonyID string, targetExecutorIDs []string) *core.Process {
	return core.CreateProcess(CreateTestFunctionSpecWithTargets(colonyID, targetExecutorIDs))
}

func CreateTestFunctionSpecWithTargets(colonyID string, targetExecutorIDs []string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, targetExecutorIDs, "test_executor_type", 1000, 100, 1, make(map[string]string), []string{}, 1, "test_label")
}

func CreateTestExecutor(colonyID string) *core.Executor {
	executor := core.CreateExecutor(core.GenerateRandomID(), "test_executor_type", core.GenerateRandomID(), colonyID, time.Now(), time.Now())
	location := core.Location{Long: 1.0, Lat: 2.0, Description: "test_desc"}
	gpu := core.GPU{Name: "test_name1", Count: 1}
	hw := core.Hardware{Model: "test_model", CPU: "test_cpu", Memory: "test_mem", Storage: "test_storage", GPU: gpu}
	sw := core.Software{Name: "test_name1", Type: "test_type1", Version: "test_version1"}
	capabilities := core.Capabilities{Hardware: hw, Software: sw}
	executor.Location = location
	executor.Capabilities = capabilities

	return executor
}

func CreateTestExecutorWithType(colonyID string, executorType string) *core.Executor {
	return core.CreateExecutor(core.GenerateRandomID(), executorType, core.GenerateRandomID(), colonyID, time.Now(), time.Now())
}

func CreateTestExecutorWithID(colonyID string, executorID string) *core.Executor {
	return core.CreateExecutor(executorID, "test_executor_type", core.GenerateRandomID(), colonyID, time.Now(), time.Now())
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

	return core.CreateExecutor(executorID, "test_executor_type", core.GenerateRandomID(), colonyID, time.Now(), time.Now()), executorPrvKey, nil
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
	funcSpec1 := CreateTestFunctionSpec(colonyID)
	funcSpec1.NodeName = "task1"
	funcSpec2 := CreateTestFunctionSpec(colonyID)
	funcSpec2.NodeName = "task2"
	funcSpec2.AddDependency("task1")
	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10, -1)
	return generator
}

func FakeGeneratorSingleProcess(t *testing.T, colonyID string) *core.Generator {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec1 := CreateTestFunctionSpec(colonyID)
	funcSpec1.NodeName = "task1"
	workflowSpec.AddFunctionSpec(funcSpec1)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10, -1)
	return generator
}

func FakeCron(t *testing.T, colonyID string) *core.Cron {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec1 := CreateTestFunctionSpec(colonyID)
	funcSpec1.NodeName = "task1"
	funcSpec2 := CreateTestFunctionSpec(colonyID)
	funcSpec2.NodeName = "task2"
	funcSpec2.AddDependency("task1")
	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	return cron
}

func FakeSingleCron(t *testing.T, colonyID string) *core.Cron {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec := CreateTestFunctionSpec(colonyID)
	funcSpec.NodeName = "task1"
	workflowSpec.AddFunctionSpec(funcSpec)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	return cron
}
