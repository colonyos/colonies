package server

import (
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestVerifyWorkflowSpec(t *testing.T) {
	colonyID := core.GenerateRandomID()

	processSpec1 := core.ProcessSpec{
		Name:        "gen_task1",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisworker"},
		Env:         make(map[string]string)}

	processSpec2 := core.ProcessSpec{
		Name:        "gen_task2",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisworker"},
		Env:         make(map[string]string)}

	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec2.AddDependency("task1")
	workflowSpec.AddProcessSpec(&processSpec1)
	workflowSpec.AddProcessSpec(&processSpec2)

	err := VerifyWorkflowSpec(workflowSpec) // Should not work
	assert.NotNil(t, err)

	processSpec1 = core.ProcessSpec{
		Name:        "gen_task1",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisworker"},
		Env:         make(map[string]string)}

	processSpec2 = core.ProcessSpec{
		Name:        "gen_task2",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisworker"},
		Env:         make(map[string]string)}

	workflowSpec = core.CreateWorkflowSpec(colonyID)
	processSpec2.AddDependency("gen_task1") // Should work
	workflowSpec.AddProcessSpec(&processSpec1)
	workflowSpec.AddProcessSpec(&processSpec2)

	err = VerifyWorkflowSpec(workflowSpec) // Should not work
	assert.Nil(t, err)
}
