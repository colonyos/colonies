package scheduler

import (
	"colonies/pkg/core"
	"fmt"
	"sort"
)

type bySubmissionTime []*core.Task

func (c bySubmissionTime) Len() int {
	return len(c)
}

func (c bySubmissionTime) Less(i, j int) bool {
	return c[i].SubmissionTime().UnixNano() > c[j].SubmissionTime().UnixNano()
}

func (c bySubmissionTime) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type BasicScheduler struct {
}

func CreateBasicScheduler() *BasicScheduler {
	return &BasicScheduler{}
}

func (scheduler *BasicScheduler) printCandidates(candidates []*core.Task) {
	for _, c := range candidates {
		fmt.Println(c.TargetColonyID())
		fmt.Println(c.SubmissionTime())
	}
}

func (scheduler *BasicScheduler) Select(workerID string, candidates []*core.Task) *core.Task {
	if len(candidates) == 0 {
		return nil
	}

	// First, check if there is task candidate target this specific worker
	for _, candidate := range candidates {
		for _, targetWorkerID := range candidate.TargetWorkerIDs() {
			if targetWorkerID == workerID {
				return candidate
			}
		}
	}

	// Ok, let's look for any task and pick the oldest task
	c := bySubmissionTime(candidates)

	sort.Sort(&c)

	return c[0]
}
