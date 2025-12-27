package database

import "github.com/colonyos/colonies/pkg/core"

type ProcessDatabase interface {
	AddProcess(process *core.Process) error
	GetProcesses() ([]*core.Process, error)
	GetProcessByID(processID string) (*core.Process, error)
	FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error)
	FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)
	FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error)
	FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error)
	FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error)
	FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error)
	FindAllRunningProcesses() ([]*core.Process, error)
	FindAllWaitingProcesses() ([]*core.Process, error)
	FindCandidates(colonyName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error)
	FindCandidatesByName(colonyName string, executorName string, executorType string, executorLocationName string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error)
	RemoveProcessByID(processID string) error
	RemoveAllProcesses() error
	RemoveAllWaitingProcessesByColonyName(colonyName string) error
	RemoveAllRunningProcessesByColonyName(colonyName string) error
	RemoveAllSuccessfulProcessesByColonyName(colonyName string) error
	RemoveAllFailedProcessesByColonyName(colonyName string) error
	RemoveAllProcessesByColonyName(colonyName string) error
	RemoveAllProcessesByProcessGraphID(processGraphID string) error
	RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error
	ResetProcess(process *core.Process) error
	SetInput(processID string, output []interface{}) error
	SetOutput(processID string, output []interface{}) error
	SetErrors(processID string, errs []string) error
	SetProcessState(processID string, state int) error
	SetParents(processID string, parents []string) error
	SetChildren(processID string, children []string) error
	SetWaitForParents(processID string, waitingForParent bool) error
	Assign(executorID string, process *core.Process) error
	SelectAndAssign(colonyName string, executorID string, executorName string, executorType string, executorLocation string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) (*core.Process, error)
	Unassign(process *core.Process) error
	MarkSuccessful(processID string) (float64, float64, error)
	MarkFailed(processID string, errs []string) error
	CountProcesses() (int, error)
	CountWaitingProcesses() (int, error)
	CountRunningProcesses() (int, error)
	CountSuccessfulProcesses() (int, error)
	CountFailedProcesses() (int, error)
	CountWaitingProcessesByColonyName(colonyName string) (int, error)
	CountRunningProcessesByColonyName(colonyName string) (int, error)
	CountSuccessfulProcessesByColonyName(colonyName string) (int, error)
	CountFailedProcessesByColonyName(colonyName string) (int, error)
}