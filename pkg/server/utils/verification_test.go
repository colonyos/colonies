package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestVerifyFunctionSpecValidPriority(t *testing.T) {
	funcSpec := &core.FunctionSpec{
		Priority: 0,
	}
	err := VerifyFunctionSpec(funcSpec)
	assert.Nil(t, err)

	funcSpec.Priority = constants.MAX_PRIORITY
	err = VerifyFunctionSpec(funcSpec)
	assert.Nil(t, err)

	funcSpec.Priority = constants.MIN_PRIORITY
	err = VerifyFunctionSpec(funcSpec)
	assert.Nil(t, err)
}

func TestVerifyFunctionSpecInvalidPriority(t *testing.T) {
	funcSpec := &core.FunctionSpec{
		Priority: constants.MIN_PRIORITY - 1,
	}
	err := VerifyFunctionSpec(funcSpec)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "priority outside range")

	funcSpec.Priority = constants.MAX_PRIORITY + 1
	err = VerifyFunctionSpec(funcSpec)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "priority outside range")
}

func TestVerifyWorkflowSpec(t *testing.T) {
	colonyName := core.GenerateRandomID()

	argsif := make([]interface{}, 1)
	argsif[0] = "arg1"

	kwargsif := make(map[string]interface{}, 1)
	kwargsif["name"] = "arg1"

	funcSpec1 := core.FunctionSpec{
		NodeName:    "gen_task1",
		FuncName:    "gen_test_func",
		Args:        argsif,
		KwArgs:      kwargsif,
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyName: colonyName, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	funcSpec2 := core.FunctionSpec{
		NodeName:    "gen_task2",
		FuncName:    "gen_test_func",
		Args:        argsif,
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyName: colonyName, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	workflowSpec := core.CreateWorkflowSpec(colonyName)
	funcSpec2.AddDependency("task1")
	workflowSpec.AddFunctionSpec(&funcSpec1)
	workflowSpec.AddFunctionSpec(&funcSpec2)

	err := VerifyWorkflowSpec(workflowSpec) // Should not work
	assert.NotNil(t, err)

	funcSpec1 = core.FunctionSpec{
		NodeName:    "gen_task1",
		FuncName:    "gen_test_func",
		Args:        argsif,
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyName: colonyName, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	funcSpec2 = core.FunctionSpec{
		NodeName:    "gen_task2",
		FuncName:    "gen_test_func",
		Args:        argsif,
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyName: colonyName, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	workflowSpec = core.CreateWorkflowSpec(colonyName)
	funcSpec2.AddDependency("gen_task1") // Should work
	workflowSpec.AddFunctionSpec(&funcSpec1)
	workflowSpec.AddFunctionSpec(&funcSpec2)

	err = VerifyWorkflowSpec(workflowSpec) // Should not work
	assert.Nil(t, err)
}
