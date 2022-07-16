package planner

import "github.com/colonyos/colonies/pkg/core"

type Planner interface {
	Select(runtimeID string, candidates []*core.Process, latest bool) (*core.Process, error)
	Prioritize(runtimeID string, candidates []*core.Process, count int, latest bool) []*core.Process
}
