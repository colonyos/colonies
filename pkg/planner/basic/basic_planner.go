package basic

import (
	"errors"
	"fmt"
	"sort"

	"github.com/colonyos/colonies/pkg/core"
)

type byOldestSubmissionTime []*core.Process

func (c byOldestSubmissionTime) Len() int {
	return len(c)
}

func (c byOldestSubmissionTime) Less(i, j int) bool {
	return c[i].PriorityTime < c[j].PriorityTime
}

func (c byOldestSubmissionTime) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type byLatestSubmissionTime []*core.Process

func (c byLatestSubmissionTime) Len() int {
	return len(c)
}

func (c byLatestSubmissionTime) Less(i, j int) bool {
	return c[i].PriorityTime > c[j].PriorityTime
}

func (c byLatestSubmissionTime) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type BasicPlanner struct {
}

func CreatePlanner() *BasicPlanner {
	return &BasicPlanner{}
}

func (planner *BasicPlanner) printCandidates(candidates []*core.Process) {
	for _, c := range candidates {
		fmt.Println(c.ID)
		fmt.Println(c.SubmissionTime)
	}
}

func (planner *BasicPlanner) Select(executorID string, candidates []*core.Process, latest bool) (*core.Process, error) {
	prioritizedProcesses := planner.Prioritize(executorID, candidates, 1, latest)
	if len(prioritizedProcesses) < 1 {
		return nil, errors.New("No processes can be selected for executor with Id <" + executorID + ">")
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

func (planner *BasicPlanner) Prioritize(executorID string, candidates []*core.Process, count int, latest bool) []*core.Process {
	var prioritizedCandidates []*core.Process
	if len(candidates) == 0 {
		return prioritizedCandidates
	}

	// First, check if there is process candidate target this specific executor
	for _, candidate := range candidates {
		if len(candidate.FunctionSpec.Conditions.ExecutorIDs) == 0 {
			prioritizedCandidates = append(prioritizedCandidates, candidate)
		} else {
			for _, targetExecutorID := range candidate.FunctionSpec.Conditions.ExecutorIDs {
				if targetExecutorID == executorID {
					prioritizedCandidates = append(prioritizedCandidates, candidate)
				}
			}
		}
	}

	if latest {
		c := byLatestSubmissionTime(prioritizedCandidates)
		sort.Sort(&c)
		return c[:min(count, len(prioritizedCandidates))]
	} else {
		c := byOldestSubmissionTime(prioritizedCandidates)
		sort.Sort(&c)
		return c[:min(count, len(prioritizedCandidates))]
	}
}
