package basic

import (
	"colonies/pkg/core"
	"errors"
	"fmt"
	"sort"
)

type bySubmissionTime []*core.Process

func (c bySubmissionTime) Len() int {
	return len(c)
}

func (c bySubmissionTime) Less(i, j int) bool {
	return c[i].SubmissionTime.UnixNano() < c[j].SubmissionTime.UnixNano()
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
		fmt.Println(c.ID)
		fmt.Println(c.SubmissionTime)
	}
}

func (scheduler *BasicScheduler) Select(runtimeID string, candidates []*core.Process) (*core.Process, error) {
	prioritizedProcesses := scheduler.Prioritize(runtimeID, candidates, 1)
	if len(prioritizedProcesses) < 1 {
		return nil, errors.New("No processes can be selected")
	}

	return prioritizedProcesses[0], nil
}

// There is no built-in max or min function for integers, but it’s simple to write
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (scheduler *BasicScheduler) Prioritize(runtimeID string, candidates []*core.Process, count int) []*core.Process {
	var prioritizedCandidates []*core.Process
	if len(candidates) == 0 {
		return prioritizedCandidates
	}

	// First, check if there is process candidate target this specific runtime
	for _, candidate := range candidates {
		if len(candidate.ProcessSpec.Conditions.RuntimeIDs) == 0 {
			prioritizedCandidates = append(prioritizedCandidates, candidate)
		} else {
			for _, targetRuntimeID := range candidate.ProcessSpec.Conditions.RuntimeIDs {
				if targetRuntimeID == runtimeID {
					prioritizedCandidates = append(prioritizedCandidates, candidate)
				}
			}
		}
	}

	// Ok, let's look for any process and pick the oldest process
	c := bySubmissionTime(prioritizedCandidates)
	sort.Sort(&c)

	return c[:min(count, len(prioritizedCandidates))]
}
