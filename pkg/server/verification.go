package server

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func VerifyWorkflowSpec(workflowSpec *core.WorkflowSpec) error {
	processMap := make(map[string]*core.Process)
	for _, funcSpec := range workflowSpec.FunctionSpecs {
		process := core.CreateProcess(&funcSpec)
		processMap[process.FunctionSpec.NodeName] = process
	}

	for _, process := range processMap {
		for _, dependsOn := range process.FunctionSpec.Conditions.Dependencies {
			parentProcess := processMap[dependsOn]
			if parentProcess == nil {
				msg := "Failed to submit workflow, invalid dependencies, are you depending on a nodename that does not exits?"
				return errors.New(msg)
			}
		}
	}

	return nil
}
