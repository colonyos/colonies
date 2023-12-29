package database

import (
	"github.com/colonyos/colonies/pkg/core"
)

type ProcessLookup interface {
	FindCandidates(colonyName string, executorType string, count int) ([]*core.Process, error)
	FindCandidatesByName(colonyName string, executorName string, executorType string, count int) ([]*core.Process, error)
}