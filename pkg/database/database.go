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

	// Colony functions ...
	AddColony(colony *core.Colony) error
	GetColonies() ([]*core.Colony, error)
	GetColonyByID(id string) (*core.Colony, error)
	RenameColony(id string, name string) error
	DeleteColonyByID(colonyID string) error
	CountColonies() (int, error)

	// Executor functions ...
	AddExecutor(executor *core.Executor) error
	AddOrReplaceExecutor(executor *core.Executor) error
	GetExecutors() ([]*core.Executor, error)
	GetExecutorByID(executorID string) (*core.Executor, error)
	GetExecutorsByColonyID(colonyID string) ([]*core.Executor, error)
	GetExecutorByName(colonyID string, executorName string) (*core.Executor, error)
	ApproveExecutor(executor *core.Executor) error
	RejectExecutor(executor *core.Executor) error
	MarkAlive(executor *core.Executor) error
	DeleteExecutorByID(executorID string) error
	DeleteExecutorsByColonyID(colonyID string) error
	CountExecutors() (int, error)
	CountExecutorsByColonyID(colonyID string) (int, error)

	// Function functions ...
	AddFunction(function *core.Function) error
	GetFunctionByID(functionID string) (*core.Function, error)
	GetFunctionsByExecutorID(executorID string) ([]*core.Function, error)
	GetFunctionsByColonyID(colonyID string) ([]*core.Function, error)
	GetFunctionsByExecutorIDAndName(executorID string, name string) (*core.Function, error)
	UpdateFunctionStats(executorID string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error
	DeleteFunctionByID(functionID string) error
	DeleteFunctionByName(executorID string, name string) error
	DeleteFunctionsByExecutorID(executorID string) error
	DeleteFunctionsByColonyID(colonyID string) error
	DeleteFunctions() error

	// Process functions ...
	AddProcess(process *core.Process) error
	GetProcesses() ([]*core.Process, error)
	GetProcessByID(processID string) (*core.Process, error)
	FindProcessesByColonyID(colonyID string, seconds int, state int) ([]*core.Process, error)
	FindProcessesByExecutorID(colonyID string, executorID string, seconds int, state int) ([]*core.Process, error)
	FindWaitingProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	FindRunningProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	FindSuccessfulProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	FindFailedProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	FindAllRunningProcesses() ([]*core.Process, error)
	FindAllWaitingProcesses() ([]*core.Process, error)
	FindUnassignedProcesses(colonyID string, executorID string, executorType string, count int) ([]*core.Process, error)
	DeleteProcessByID(processID string) error
	DeleteAllProcesses() error
	DeleteAllWaitingProcessesByColonyID(colonyID string) error
	DeleteAllRunningProcessesByColonyID(colonyID string) error
	DeleteAllSuccessfulProcessesByColonyID(colonyID string) error
	DeleteAllFailedProcessesByColonyID(colonyID string) error
	DeleteAllProcessesByColonyID(colonyID string) error
	DeleteAllProcessesByProcessGraphID(processGraphID string) error
	DeleteAllProcessesInProcessGraphsByColonyID(colonyID string) error
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
	CountWaitingProcessesByColonyID(colonyID string) (int, error)
	CountRunningProcessesByColonyID(colonyID string) (int, error)
	CountSuccessfulProcessesByColonyID(colonyID string) (int, error)
	CountFailedProcessesByColonyID(colonyID string) (int, error)

	// Attribute functions
	AddAttribute(attribute core.Attribute) error
	AddAttributes(attribute []core.Attribute) error
	GetAttributeByID(attributeID string) (core.Attribute, error)
	GetAttributesByColonyID(colonyID string) ([]core.Attribute, error)
	GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error)
	GetAttributes(targetID string) ([]core.Attribute, error)
	GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error)
	UpdateAttribute(attribute core.Attribute) error
	DeleteAttributeByID(attributeID string) error
	DeleteAllAttributesByColonyID(colonyID string) error
	DeleteAllAttributesByColonyIDWithState(colonyID string, state int) error
	DeleteAllAttributesByProcessGraphID(processGraphID string) error
	DeleteAllAttributesInProcessGraphsByColonyID(colonyID string) error
	DeleteAllAttributesInProcessGraphsByColonyIDWithState(colonyID string, state int) error
	DeleteAttributesByTargetID(targetID string, attributeType int) error
	DeleteAllAttributesByTargetID(targetID string) error
	DeleteAllAttributes() error

	// ProcessGraph functions
	AddProcessGraph(processGraph *core.ProcessGraph) error
	GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	SetProcessGraphState(processGraphID string, state int) error
	FindWaitingProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	FindRunningProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	FindSuccessfulProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	FindFailedProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	DeleteProcessGraphByID(processGraphID string) error
	DeleteAllProcessGraphsByColonyID(colonyID string) error
	DeleteAllWaitingProcessGraphsByColonyID(colonyID string) error
	DeleteAllRunningProcessGraphsByColonyID(colonyID string) error
	DeleteAllSuccessfulProcessGraphsByColonyID(colonyID string) error
	DeleteAllFailedProcessGraphsByColonyID(colonyID string) error
	CountWaitingProcessGraphs() (int, error)
	CountRunningProcessGraphs() (int, error)
	CountSuccessfulProcessGraphs() (int, error)
	CountFailedProcessGraphs() (int, error)
	CountWaitingProcessGraphsByColonyID(colonyID string) (int, error)
	CountRunningProcessGraphsByColonyID(colonyID string) (int, error)
	CountSuccessfulProcessGraphsByColonyID(colonyID string) (int, error)
	CountFailedProcessGraphsByColonyID(colonyID string) (int, error)

	// Generator functions
	AddGenerator(generator *core.Generator) error
	SetGeneratorLastRun(generatorID string) error
	SetGeneratorFirstPack(generatorID string) error
	GetGeneratorByID(generatorID string) (*core.Generator, error)
	GetGeneratorByName(name string) (*core.Generator, error)
	FindGeneratorsByColonyID(colonyID string, count int) ([]*core.Generator, error)
	FindAllGenerators() ([]*core.Generator, error)
	DeleteGeneratorByID(generatorID string) error
	DeleteAllGeneratorsByColonyID(colonyID string) error

	// Generator args functions
	AddGeneratorArg(generatorArg *core.GeneratorArg) error
	GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error)
	CountGeneratorArgs(generatorID string) (int, error)
	DeleteGeneratorArgByID(generatorArgsID string) error
	DeleteAllGeneratorArgsByGeneratorID(generatorID string) error
	DeleteAllGeneratorArgsByColonyID(generatorID string) error

	// Cron functions
	AddCron(cron *core.Cron) error
	UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error
	GetCronByID(cronID string) (*core.Cron, error)
	FindCronsByColonyID(colonyID string, count int) ([]*core.Cron, error)
	FindAllCrons() ([]*core.Cron, error)
	DeleteCronByID(cronID string) error
	DeleteAllCronsByColonyID(colonyID string) error

	// Distributed locking
	Lock(timeout int) error
	Unlock() error

	// Retention management
	ApplyRetentionPolicy(retentionPeriod int64) error

	// Logging
	AddLog(processID string, colonyID string, executorID string, msg string) error
	GetLogsByProcessID(processID string, limit int) ([]core.Log, error)
}
