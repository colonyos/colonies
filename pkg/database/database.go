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
	DeleteColonyByID(colonyID string) error
	CountColonies() (int, error)

	// Runtime functions ...
	AddRuntime(runtime *core.Runtime) error
	GetRuntimes() ([]*core.Runtime, error)
	GetRuntimeByID(runtimeID string) (*core.Runtime, error)
	GetRuntimesByColonyID(colonyID string) ([]*core.Runtime, error)
	ApproveRuntime(runtime *core.Runtime) error
	RejectRuntime(runtime *core.Runtime) error
	MarkAlive(runtime *core.Runtime) error
	DeleteRuntimeByID(runtimeID string) error
	DeleteRuntimesByColonyID(colonyID string) error
	CountRuntimes() (int, error)
	CountRuntimesByColonyID(colonyID string) (int, error)

	// Process functions ...
	AddProcess(process *core.Process) error
	GetProcesses() ([]*core.Process, error)
	GetProcessByID(processID string) (*core.Process, error)
	FindProcessesByColonyID(colonyID string, seconds int, state int) ([]*core.Process, error)
	FindProcessesByRuntimeID(colonyID string, runtimeID string, seconds int, state int) ([]*core.Process, error)
	FindWaitingProcesses(colonyID string, count int) ([]*core.Process, error)
	FindRunningProcesses(colonyID string, count int) ([]*core.Process, error)
	FindAllRunningProcesses() ([]*core.Process, error)
	FindAllWaitingProcesses() ([]*core.Process, error)
	FindSuccessfulProcesses(colonyID string, count int) ([]*core.Process, error)
	FindFailedProcesses(colonyID string, count int) ([]*core.Process, error)
	FindUnassignedProcesses(colonyID string, runtimeID string, runtimeType string, count int, latest bool) ([]*core.Process, error)
	DeleteProcessByID(processID string) error
	DeleteAllProcesses() error
	DeleteAllProcessesByColonyID(colonyID string) error
	DeleteAllProcessesByProcessGraphID(processGraphID string) error
	DeleteAllProcessesInProcessGraphsByColonyID(colonyID string) error
	ResetProcess(process *core.Process) error
	SetInput(processID string, output []string) error
	SetOutput(processID string, output []string) error
	SetErrors(processID string, errs []string) error
	SetProcessState(processID string, state int) error
	SetParents(processID string, parents []string) error
	SetChildren(processID string, children []string) error
	SetWaitForParents(processID string, waitingForParent bool) error
	ResetAllProcesses(process *core.Process) error
	AssignRuntime(runtimeID string, process *core.Process) error
	UnassignRuntime(process *core.Process) error
	MarkSuccessful(processID string) error
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
	GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error)
	GetAttributes(targetID string) ([]core.Attribute, error)
	GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error)
	UpdateAttribute(attribute core.Attribute) error
	DeleteAttributeByID(attributeID string) error
	DeleteAllAttributesByColonyID(colonyID string) error
	DeleteAllAttributesByProcessGraphID(processGraphID string) error
	DeleteAllAttributesInProcessGraphsByColonyID(colonyID string) error
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
}
