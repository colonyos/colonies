package database

import "github.com/colonyos/colonies/pkg/core"

type ProcessGraphDatabase interface {
	AddProcessGraph(processGraph *core.ProcessGraph) error
	GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	SetProcessGraphState(processGraphID string, state int) error
	FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	RemoveProcessGraphByID(processGraphID string) error
	RemoveAllProcessGraphsByColonyName(colonyName string) error
	RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error
	RemoveAllRunningProcessGraphsByColonyName(colonyName string) error
	RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error
	RemoveAllFailedProcessGraphsByColonyName(colonyName string) error
	CountWaitingProcessGraphs() (int, error)
	CountRunningProcessGraphs() (int, error)
	CountSuccessfulProcessGraphs() (int, error)
	CountFailedProcessGraphs() (int, error)
	CountWaitingProcessGraphsByColonyName(colonyName string) (int, error)
	CountRunningProcessGraphsByColonyName(colonyName string) (int, error)
	CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error)
	CountFailedProcessGraphsByColonyName(colonyName string) (int, error)
}