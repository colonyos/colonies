package scheduler

import "github.com/colonyos/colonies/pkg/core"

type Scheduler interface {
	Select(runtimeID string, candidates []*core.Process) (*core.Process, error)
	Prioritize(runtimeID string, candidates []*core.Process, count int) []*core.Process
}
