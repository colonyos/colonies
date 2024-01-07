package database

import (
	"time"

	"github.com/colonyos/colonies/pkg/core"
)

type Database interface {
	// General
	Close()
	Initialize() error
	Drop() error

	// Users functions ...
	AddUser(user *core.User) error
	GetUsersByColonyName(colonyName string) ([]*core.User, error)
	GetUserByID(colonyName string, userID string) (*core.User, error)
	GetUserByName(colonyName string, name string) (*core.User, error)
	RemoveUserByID(colonyName string, userID string) error
	RemoveUserByName(colonyName string, name string) error
	RemoveUsersByColonyName(colonyName string) error

	// Colony functions ...
	AddColony(colony *core.Colony) error
	GetColonies() ([]*core.Colony, error)
	GetColonyByID(id string) (*core.Colony, error)
	GetColonyByName(name string) (*core.Colony, error)
	RenameColony(colonyName string, newColonyName string) error
	RemoveColonyByName(colonyName string) error
	CountColonies() (int, error)

	// Executor functions ...
	AddExecutor(executor *core.Executor) error
	SetAllocations(colonyName string, executorName string, allocations core.Allocations) error
	GetExecutors() ([]*core.Executor, error)
	GetExecutorByID(executorID string) (*core.Executor, error)
	GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error)
	GetExecutorByName(colonyName string, executorName string) (*core.Executor, error)
	ApproveExecutor(executor *core.Executor) error
	RejectExecutor(executor *core.Executor) error
	MarkAlive(executor *core.Executor) error
	RemoveExecutorByName(colonyName string, executorName string) error
	RemoveExecutorsByColonyName(colonyName string) error
	CountExecutors() (int, error)
	CountExecutorsByColonyName(colonyName string) (int, error)

	// Function functions ...
	AddFunction(function *core.Function) error
	GetFunctionByID(functionID string) (*core.Function, error)
	GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error)
	GetFunctionsByColonyName(colonyName string) ([]*core.Function, error)
	GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error)
	UpdateFunctionStats(colonyName string, executorName string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error
	RemoveFunctionByID(functionID string) error
	RemoveFunctionByName(colonyName string, executorName string, name string) error
	RemoveFunctionsByExecutorName(colonyName string, executorName string) error
	RemoveFunctionsByColonyName(colonyName string) error
	RemoveFunctions() error

	// Process functions ...
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
	FindCandidates(colonyName string, executorType string, cpu int64, memory int64, gpuName string, gpuMem int64, gpuCount int, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error)
	FindCandidatesByName(colonyName string, executorName string, executorType string, cpu int64, memory int64, gpuName string, gpuMem int64, gpuCount int, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error)
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

	// Attribute functions
	AddAttribute(attribute core.Attribute) error
	AddAttributes(attribute []core.Attribute) error
	GetAttributeByID(attributeID string) (core.Attribute, error)
	GetAttributesByColonyName(colonyName string) ([]core.Attribute, error)
	GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error)
	GetAttributes(targetID string) ([]core.Attribute, error)
	GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error)
	UpdateAttribute(attribute core.Attribute) error
	RemoveAttributeByID(attributeID string) error
	RemoveAllAttributesByColonyName(colonyName string) error
	RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error
	RemoveAllAttributesByProcessGraphID(processGraphID string) error
	RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error
	RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error
	RemoveAttributesByTargetID(targetID string, attributeType int) error
	RemoveAllAttributesByTargetID(targetID string) error
	RemoveAllAttributes() error

	// ProcessGraph functions
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

	// Generator functions
	AddGenerator(generator *core.Generator) error
	SetGeneratorLastRun(generatorID string) error
	SetGeneratorFirstPack(generatorID string) error
	GetGeneratorByID(generatorID string) (*core.Generator, error)
	GetGeneratorByName(colonyName string, name string) (*core.Generator, error)
	FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error)
	FindAllGenerators() ([]*core.Generator, error)
	RemoveGeneratorByID(generatorID string) error
	RemoveAllGeneratorsByColonyName(colonyName string) error

	// Generator args functions
	AddGeneratorArg(generatorArg *core.GeneratorArg) error
	GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error)
	CountGeneratorArgs(generatorID string) (int, error)
	RemoveGeneratorArgByID(generatorArgsID string) error
	RemoveAllGeneratorArgsByGeneratorID(generatorID string) error
	RemoveAllGeneratorArgsByColonyName(generatorID string) error

	// Cron functions
	AddCron(cron *core.Cron) error
	UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error
	GetCronByID(cronID string) (*core.Cron, error)
	GetCronByName(colonyName string, cronName string) (*core.Cron, error)
	FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error)
	FindAllCrons() ([]*core.Cron, error)
	RemoveCronByID(cronID string) error
	RemoveAllCronsByColonyName(colonyName string) error

	// Distributed locking
	Lock(timeout int) error
	Unlock() error

	// Retention management
	ApplyRetentionPolicy(retentionPeriod int64) error

	// Logging
	AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error
	GetLogsByProcessID(processID string, limit int) ([]*core.Log, error)
	GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error)
	GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error)
	GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error)
	RemoveLogsByColonyName(colonyName string) error
	CountLogs(colonyName string) (int, error)
	SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error)

	// File management
	AddFile(file *core.File) error
	GetFileByID(colonyName string, fileID string) (*core.File, error)
	GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error)
	GetFileByName(colonyName string, label string, name string) ([]*core.File, error)
	GetFilenamesByLabel(colonyName string, label string) ([]string, error)
	GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error)
	RemoveFileByID(colonyName string, fileID string) error
	RemoveFileByName(colonyName string, label string, name string) error
	GetFileLabels(colonyName string) ([]*core.Label, error)
	GetFileLabelsByName(colonyName string, name string) ([]*core.Label, error)
	CountFilesWithLabel(colonyName string, label string) (int, error)
	CountFiles(colonyName string) (int, error)

	// Snapshots
	CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error)
	GetSnapshotByID(colonNamey string, snapshotID string) (*core.Snapshot, error)
	GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error)
	RemoveSnapshotByID(colonyName string, snapshotID string) error
	GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error)
	RemoveSnapshotByName(colonyName string, name string) error
	RemoveSnapshotsByColonyName(colonyName string) error

	// Security
	SetServerID(oldServerID, newServerID string) error
	GetServerID() (string, error)
	ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error
	ChangeUserID(colonyName string, oldUserID, newUserID string) error
	ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error
}
