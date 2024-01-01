package utils

import (
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/stretchr/testify/assert"
)

func CreateTestUser(colonyID string, name string) *core.User {
	userID := core.GenerateRandomID()
	email := "test@test.com"
	phone := "12345677"
	return core.CreateUser(colonyID, userID, name, email, phone)
}

func CreateTestUserWithKey(colonyID string, name string) (*core.User, string, error) {
	crypto := crypto.CreateCrypto()
	userPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	userID, err := crypto.GenerateID(userPrvKey)
	if err != nil {
		return nil, "", err
	}

	email := "test@test.com"
	phone := "12345677"
	return core.CreateUser(colonyID, userID, name, email, phone), userPrvKey, nil
}

func CreateTestProcess(colonyName string) *core.Process {
	process := core.CreateProcess(CreateTestFunctionSpec(colonyName))
	process.InitiatorID = "test_initiator_id"
	process.InitiatorName = "test_initiator_name"

	return process
}

func CreateTestFunctionSpec(colonyName string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyName, []string{}, "test_executor_type", 1000, 100, 1, make(map[string]string), []string{}, 1, "test_label")
}

func CreateTestFileWithID(id string, colonyName string, now time.Time) *core.File {
	s3Object := core.S3Object{
		Server:        "test_server",
		Port:          1111,
		TLS:           true,
		AccessKey:     "test_accesskey",
		SecretKey:     "test_secretkey",
		Region:        "test_region",
		EncryptionKey: "test_encrytionkey",
		EncryptionAlg: "test_encrytionalg",
		Object:        "test_object",
		Bucket:        "test_bucket",
	}
	ref := core.Reference{Protocol: "s3", S3Object: s3Object}
	file := core.File{
		ID:          id,
		ColonyName:  colonyName,
		Label:       "test_label",
		Name:        "test_name",
		Size:        1111,
		Checksum:    "test_checksum",
		ChecksumAlg: "test_checksumalg",
		Reference:   ref,
		Added:       now}

	return &file
}

func CreateTestFile(colonyName string) *core.File {
	s3Object := core.S3Object{
		Server:        "test_server",
		Port:          1111,
		TLS:           true,
		AccessKey:     "test_accesskey",
		SecretKey:     "test_secretkey",
		Region:        "test_region",
		EncryptionKey: "test_encrytionkey",
		EncryptionAlg: "test_encrytionalg",
		Object:        "test_object",
		Bucket:        "test_bucket",
	}
	ref := core.Reference{Protocol: "s3", S3Object: s3Object}
	file := core.File{
		ID:          "",
		ColonyName:  colonyName,
		Label:       "test_label",
		Name:        "test_name",
		Size:        1111,
		Checksum:    "test_checksum",
		ChecksumAlg: "test_checksumalg",
		Reference:   ref,
		Added:       time.Time{}}

	return &file
}

func CreateTestProcessWithType(colonyID string, executorType string) *core.Process {
	process := core.CreateProcess(CreateTestFunctionSpecWithType(colonyID, executorType))
	process.InitiatorID = "test_initiator_id"
	process.InitiatorName = "test_initiator_name"

	return process
}

func CreateTestFunctionSpecWithType(colonyID string, executorType string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, []string{}, executorType, 1000, 100, 1, make(map[string]string), []string{}, 1, "test_label")
}

func CreateTestProcessWithEnv(colonyID string, env map[string]string) *core.Process {
	process := core.CreateProcess(CreateTestFunctionSpecWithEnv(colonyID, env))
	process.InitiatorID = "test_initiator_id"
	process.InitiatorName = "test_initiator_name"

	return process
}

func CreateTestFunctionSpecWithEnv(colonyID string, env map[string]string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, []string{}, "test_executor_type", 1000, 100, 1, env, []string{}, 1, "test_label")
}

func CreateTestProcessWithTargets(colonyID string, targetExecutorIDs []string) *core.Process {
	process := core.CreateProcess(CreateTestFunctionSpecWithTargets(colonyID, targetExecutorIDs))
	process.InitiatorID = "test_initiator_id"
	process.InitiatorName = "test_initiator_name"

	return process
}

func CreateTestFunctionSpecWithTargets(colonyID string, targetExecutorIDs []string) *core.FunctionSpec {
	args := make([]interface{}, 1)
	args[0] = "test_arg"
	kwargs := make(map[string]interface{}, 1)
	kwargs["name"] = "test_arg"
	return core.CreateFunctionSpec("test_name", "test_func", args, kwargs, colonyID, targetExecutorIDs, "test_executor_type", 1000, 100, 1, make(map[string]string), []string{}, 1, "test_label")
}

func CreateTestExecutor(colonyName string) *core.Executor {
	executor := core.CreateExecutor(core.GenerateRandomID(), "test_executor_type", core.GenerateRandomID(), colonyName, time.Now(), time.Now())
	location := core.Location{Long: 1.0, Lat: 2.0, Description: "test_desc"}
	gpu := core.GPU{Name: "test_name1", Count: 1}
	hw := core.Hardware{Model: "test_model", CPU: "0m", Memory: "0m", Storage: "test_storage", GPU: gpu}
	sw := core.Software{Name: "test_name1", Type: "test_type1", Version: "test_version1"}
	capabilities := core.Capabilities{Hardware: hw, Software: sw}
	executor.Location = location
	executor.Capabilities = capabilities

	return executor
}

func CreateTestExecutorWithType(colonyName string, executorType string) *core.Executor {
	return core.CreateExecutor(core.GenerateRandomID(), executorType, core.GenerateRandomID(), colonyName, time.Now(), time.Now())
}

func CreateTestExecutorWithID(colonyName string, executorID string) *core.Executor {
	return core.CreateExecutor(executorID, "test_executor_type", core.GenerateRandomID(), colonyName, time.Now(), time.Now())
}

func CreateTestExecutorWithKey(colonyName string) (*core.Executor, string, error) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", err
	}

	executorID, err := crypto.GenerateID(executorPrvKey)
	if err != nil {
		return nil, "", err
	}

	return core.CreateExecutor(executorID, "test_executor_type", core.GenerateRandomID(), colonyName, time.Now(), time.Now()), executorPrvKey, nil
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
	return core.CreateColony(colonyID, "test_colony_name"+core.GenerateRandomID()), colonyPrvKey, nil
}

func FakeGenerator(t *testing.T, colonyID string, initiatorID string, initiatorName string) *core.Generator {
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
	generator.InitiatorID = initiatorID
	generator.InitiatorName = initiatorName
	return generator
}

func FakeGeneratorSingleProcess(t *testing.T, colonyID string, initiatorID string, initiatorName string) *core.Generator {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec1 := CreateTestFunctionSpec(colonyID)
	funcSpec1.NodeName = "task1"
	workflowSpec.AddFunctionSpec(funcSpec1)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10, -1)
	generator.InitiatorID = initiatorID
	generator.InitiatorName = initiatorName
	return generator
}

func FakeCron(t *testing.T, colonyID string, initiatorID string, initiatorName string) *core.Cron {
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
	cron.InitiatorID = initiatorID
	cron.InitiatorName = initiatorName
	return cron
}

func FakeSingleCron(t *testing.T, colonyID string, initiatorID string, initiatorName string) *core.Cron {
	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec := CreateTestFunctionSpec(colonyID)
	funcSpec.NodeName = "task1"
	workflowSpec.AddFunctionSpec(funcSpec)
	jsonStr, err := workflowSpec.ToJSON()
	assert.Nil(t, err)
	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	cron.InitiatorID = initiatorID
	cron.InitiatorName = initiatorName
	return cron
}
