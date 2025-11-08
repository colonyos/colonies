package server

import (
	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server/controllers"
)

// GenericServer defines the interface that handlers can use to interact with the server
// This interface abstracts away the HTTP backend implementation
type GenericServer interface {
	// HTTP Response methods
	HandleHTTPError(c backends.Context, err error, errorCode int) bool
	SendHTTPReply(c backends.Context, payloadType string, jsonString string)
	SendEmptyHTTPReply(c backends.Context, payloadType string)
	
	// Server identity
	GetServerID() (string, error)
	
	// Security and validation
	Validator() security.Validator
	
	// Database access
	UserDB() database.UserDatabase
	ExecutorDB() database.ExecutorDatabase
	ColonyDB() database.ColonyDatabase
	SecurityDB() database.SecurityDatabase
	ProcessDB() database.ProcessDatabase
	ProcessGraphDB() database.ProcessGraphDatabase
	AttributeDB() database.AttributeDatabase
	FileDB() database.FileDatabase
	LogDB() database.LogDatabase
	FunctionDB() database.FunctionDatabase
	GeneratorDB() database.GeneratorDatabase
	CronDB() database.CronDatabase
	SnapshotDB() database.SnapshotDatabase
	ServiceDB() database.ServiceDatabase

	// Controllers
	ColonyController() controllers.ColoniesController
	ProcessController() interface {
		AddProcess(process *core.Process) (*core.Process, error)
		GetProcess(processID string) (*core.Process, error)
		GetProcesses(colonyName string, count int, state int) ([]*core.Process, error)
		FindProcesses(colonyName string, processType string, label string, initiatorID string, count int, state int) ([]*core.Process, error)
		GetProcessesByExecutorID(executorID string, count int, state int) ([]*core.Process, error)
		GetProcessesByProcessGraphID(processGraphID string, count int, state int) ([]*core.Process, error)
		RemoveProcess(processID string, state int) error
		RemoveAllProcesses(colonyName string, state int) error
		RemoveAllProcessesInProcessGraphs(colonyName string, state int) error
		RemoveAllProcessesInProcessGraphsByID(processGraphID string, state int) error
		SetProcessState(processID string, state int) error
		AssignProcess(colonyName string, executorID string) (*core.Process, error)
		UnassignProcess(processID string) (*core.Process, error)
		MarkSuccessful(processID string) (*core.Process, error)
		MarkFailed(processID string, errorMsg string) (*core.Process, error)
		CloseSuccessful(processID string, output []interface{}) (*core.Process, error)
		CloseFailed(processID string, errorMsg string) (*core.Process, error)
		SetOutput(processID string, output []interface{}) (*core.Process, error)
	}
	ExecutorController() interface {
		AddExecutor(executor *core.Executor) (*core.Executor, error)
		GetExecutor(executorID string) (*core.Executor, error)
		GetExecutorByName(colonyName string, executorName string) (*core.Executor, error)
		GetExecutors(colonyName string) ([]*core.Executor, error)
		GetExecutorsWithState(colonyName string, state int) ([]*core.Executor, error)
		RemoveExecutor(executorID string) error
		ApproveExecutor(executorID string) error
		RejectExecutor(executorID string) error
		ReportAllocation(executorID string, available bool, executorType string, nodes int, cpu string, memory string, storage string, gpu string, gpuCount int) error
	}
	GeneratorController() interface {
		AddGenerator(generator *core.Generator) (*core.Generator, error)
		GetGenerator(generatorID string) (*core.Generator, error)
		ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error)
		GetGenerators(colonyName string, count int) ([]*core.Generator, error)
		PackGenerator(generatorID string, colonyName string, arg string) error
		RemoveGenerator(generatorID string) error
		GetGeneratorPeriod() int
	}
	CronController() interface {
		AddCron(cron *core.Cron) (*core.Cron, error)
		GetCron(cronID string) (*core.Cron, error)
		GetCrons(colonyName string, count int) ([]*core.Cron, error)
		RunCron(cronID string) (*core.Cron, error)
		RemoveCron(cronID string) error
		GetCronPeriod() int
	}
}