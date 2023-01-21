package planner

import "github.com/colonyos/colonies/pkg/core"

type Planner interface {
	Select(executorID string, candidates []*core.Process, latest bool) (*core.Process, error)
	Prioritize(executorID string, candidates []*core.Process, count int, latest bool) []*core.Process
}
