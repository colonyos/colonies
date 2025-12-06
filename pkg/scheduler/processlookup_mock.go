package scheduler

import (
	"github.com/colonyos/colonies/pkg/core"
)

type processLookupMock struct {
	processTable map[string]*core.Process
}

func createProcessLookupMock() *processLookupMock {
	mock := &processLookupMock{}
	mock.processTable = make(map[string]*core.Process)

	return mock
}

func (mock *processLookupMock) addProcess(process *core.Process) {
	mock.processTable[process.ID] = process
}

func (mock *processLookupMock) FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	var c []*core.Process

	for _, process := range mock.processTable {
		if process.FunctionSpec.Conditions.ColonyName == colonyName &&
			process.State == core.WAITING &&
			len(process.FunctionSpec.Conditions.ExecutorNames) == 0 &&
			process.FunctionSpec.Conditions.ExecutorType == executorType {
			// Location filter: match if process has no location OR matches executor location
			processLocation := process.FunctionSpec.Conditions.LocationName
			if processLocation == "" || processLocation == executorLocationName {
				c = append(c, process)
			}
		}
	}

	return c, nil
}

func (mock *processLookupMock) FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	var c []*core.Process

	for _, process := range mock.processTable {
		if process.FunctionSpec.Conditions.ColonyName == colonyName &&
			process.State == core.WAITING &&
			process.FunctionSpec.Conditions.ExecutorType == executorType {
			// Location filter: match if process has no location OR matches executor location
			processLocation := process.FunctionSpec.Conditions.LocationName
			if processLocation != "" && processLocation != executorLocationName {
				continue
			}
			for _, n := range process.FunctionSpec.Conditions.ExecutorNames {
				if n == executorName {
					c = append(c, process)
				}
			}
		}
	}

	return c, nil
}
