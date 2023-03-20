package planner

import "github.com/colonyos/colonies/pkg/core"

type Planner interface {
	Select(executorID string, candidates []*core.Process) (*core.Process, error)
	Prioritize(executorID string, candidates []*core.Process, count int) []*core.Process
}
