package planner

import "github.com/colonyos/colonies/pkg/core"

type Planner interface {
	Select(runtimeID string, candidates []*core.Process) (*core.Process, error)
	Prioritize(runtimeID string, candidates []*core.Process, count int) []*core.Process
}
