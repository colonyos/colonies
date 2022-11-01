package server

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
)

func VerifyWorkflowSpec(workflowSpec *core.WorkflowSpec) error {
	processMap := make(map[string]*core.Process)
	for _, processSpec := range workflowSpec.ProcessSpecs {
		process := core.CreateProcess(&processSpec)
		processMap[process.ProcessSpec.Name] = process
	}

	for _, process := range processMap {
		for _, dependsOn := range process.ProcessSpec.Conditions.Dependencies {
			parentProcess := processMap[dependsOn]
			if parentProcess == nil {
				msg := "Failed to submit workflow, invalid dependencies, are you depending on a process spec name that does not exits?"
				return errors.New(msg)
			}
		}
	}

	return nil
}
