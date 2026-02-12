package controllers

import (
	"time"

	"github.com/colonyos/colonies/pkg/backends"
	"github.com/colonyos/colonies/pkg/channel"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
)

type Controller interface {
	GetCronPeriod() int
	GetGeneratorPeriod() int
	GetEtcdServer() *cluster.EtcdServer
	GetEventHandler() backends.RealtimeEventHandler
	GetThisNode() cluster.Node
	SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error
	SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error
	AddProcessToDB(process *core.Process) (*core.Process, error)
	AddProcess(process *core.Process) (*core.Process, error)
	AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error)
	UpdateProcessGraph(graph *core.ProcessGraph) error
	CreateProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error)
	SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error)
	GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindCancelledProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	CancelProcess(processID string) error
	CancelProcessGraph(processGraphID string) error
	CloseSuccessful(processID string, executorID string, output []interface{}) error
	NotifyChildren(process *core.Process) error
	CloseFailed(processID string, errs []string) error
	HandleDefunctProcessgraph(processGraphID string, processID string, err error) error
	Assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error)
	DistributedAssign(executor *core.Executor, colonyName string, cpu int64, memory int64, storage int64) (*AssignResult, error)
	UnassignExecutor(processID string) error
	ResetProcess(processID string) error
	AddGenerator(generator *core.Generator) (*core.Generator, error)
	PackGenerator(generatorID string, colonyName, arg string) error
	GeneratorTriggerLoop()
	TriggerGenerators()
	SubmitWorkflow(generator *core.Generator, counter int, recoveredID string) // TODO: change name, there is also a submitWorkflowSpec()
	AddCron(cron *core.Cron) (*core.Cron, error)
	RemoveGenerator(generatorID string) error
	RunCron(cronID string) (*core.Cron, error)
	RemoveCron(cronID string) error
	CalcNextRun(cron *core.Cron) time.Time
	StartCron(cron *core.Cron)
	TriggerCrons()
	CronTriggerLoop()
	ResetDatabase() error
	PauseColonyAssignments(colonyName string) error
	ResumeColonyAssignments(colonyName string) error
	AreColonyAssignmentsPaused(colonyName string) (bool, error)
	Stop()
	IsLeader() bool
	TryBecomeLeader() bool
	TimeoutLoop()
	BlockingCmdQueueWorker()
	RetentionWorker()
	CmdQueueWorker()
	GetChannelRouter() *channel.Router
}
