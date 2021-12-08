package scheduler

import "colonies/pkg/core"

type Scheduler interface {
	Select(workerID string, candidates []*core.Task) *core.Task
}
