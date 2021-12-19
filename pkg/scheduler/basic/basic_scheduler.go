package basic

import (
	"colonies/pkg/core"
	"fmt"
	"sort"
)

type bySubmissionTime []*core.Process

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

func CreateScheduler() *BasicScheduler {
	return &BasicScheduler{}
}

func (scheduler *BasicScheduler) printCandidates(candidates []*core.Process) {
	for _, c := range candidates {
		fmt.Println(c.TargetColonyID())
		fmt.Println(c.SubmissionTime())
	}
}

func (scheduler *BasicScheduler) Select(computerID string, candidates []*core.Process) *core.Process {
	if len(candidates) == 0 {
		return nil
	}

	// First, check if there is process candidate target this specific computer
	for _, candidate := range candidates {
		for _, targetComputerID := range candidate.TargetComputerIDs() {
			if targetComputerID == computerID {
				return candidate
			}
		}
	}

	// Ok, let's look for any process and pick the oldest process
	c := bySubmissionTime(candidates)

	sort.Sort(&c)

	return c[0]
}
