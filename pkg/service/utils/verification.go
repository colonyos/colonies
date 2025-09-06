package service

import (
	"errors"
	"strconv"

	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
)

func VerifyFunctionSpec(funcSpec *core.FunctionSpec) error {
	if funcSpec.Priority < constants.MIN_PRIORITY || funcSpec.Priority > constants.MAX_PRIORITY {
		msg := "Failed to submit function spec, priority outside range [" + strconv.Itoa(constants.MIN_PRIORITY) + ", " + strconv.Itoa(constants.MAX_PRIORITY) + "]"
		return errors.New(msg)
	}

	return nil
}

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
