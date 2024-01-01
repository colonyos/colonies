package database

import (
	"github.com/colonyos/colonies/pkg/core"
)

type ProcessLookup interface {
	FindCandidates(colonyName string, executorType string, cpu int64, memory int64, gpuName string, gpuMem int64, gpuCount int, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error)
	FindCandidatesByName(colonyName string, executorName string, executorType string, cpu int64, memory int64, gpuName string, gpuMem int64, gpuCount int, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error)
}
