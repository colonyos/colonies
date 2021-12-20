package scheduler

import "colonies/pkg/core"

type Scheduler interface {
	Select(computerID string, candidates []*core.Process) (*core.Process, error)
	Prioritize(computerID string, candidates []*core.Process, count int) []*core.Process
}
