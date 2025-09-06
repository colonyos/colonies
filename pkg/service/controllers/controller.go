package controllers

import (
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	websockethandlers "github.com/colonyos/colonies/pkg/service/handlers/websocket"
	servercommunication "github.com/colonyos/colonies/pkg/service/websocket"
)

type Controller interface {
	GetCronPeriod() int
	GetGeneratorPeriod() int
	GetEtcdServer() *cluster.EtcdServer
	GetEventHandler() *servercommunication.EventHandler
	GetThisNode() cluster.Node
	SubscribeProcesses(executorID string, subscription *websockethandlers.Subscription) error
	SubscribeProcess(executorID string, subscription *websockethandlers.Subscription) error
	GetColonies() ([]*core.Colony, error)
	GetColony(colonyName string) (*core.Colony, error)
	AddColony(colony *core.Colony) (*core.Colony, error)
	RemoveColony(colonyName string) error
	AddExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error)
	GetExecutor(executorID string) (*core.Executor, error)
	GetExecutorByColonyName(colonyName string) ([]*core.Executor, error)
	AddProcessToDB(process *core.Process) (*core.Process, error)
	AddProcess(process *core.Process) (*core.Process, error)
	AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error)
	GetProcess(processID string) (*core.Process, error)
	FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)
	UpdateProcessGraph(graph *core.ProcessGraph) error
	CreateProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error)
	SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error)
	GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	RemoveProcess(processID string) error
	RemoveAllProcesses(colonyName string, state int) error
	RemoveProcessGraph(processID string) error
	RemoveAllProcessGraphs(colonyName string, state int) error
	SetOutput(processID string, output []interface{}) error
	CloseSuccessful(processID string, executorID string, output []interface{}) error
	NotifyChildren(process *core.Process) error
	CloseFailed(processID string, errs []string) error
	HandleDefunctProcessgraph(processGraphID string, processID string, err error) error
	Assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error)
	UnassignExecutor(processID string) error
	ResetProcess(processID string) error
	GetColonyStatistics(colonyName string) (*core.Statistics, error)
	GetStatistics() (*core.Statistics, error)
	AddAttribute(attribute *core.Attribute) (*core.Attribute, error)
	GetAttribute(attributeID string) (*core.Attribute, error)
	AddFunction(function *core.Function) (*core.Function, error)
	GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error)
	GetFunctionsByColonyName(colonyName string) ([]*core.Function, error)
	GetFunctionByID(functionID string) (*core.Function, error)
	RemoveFunction(functionID string) error
	AddGenerator(generator *core.Generator) (*core.Generator, error)
	GetGenerator(generatorID string) (*core.Generator, error)
	ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error)
	GetGenerators(colonyName string, count int) ([]*core.Generator, error)
	PackGenerator(generatorID string, colonyName, arg string) error
	GeneratorTriggerLoop()
	TriggerGenerators()
	SubmitWorkflow(generator *core.Generator, counter int, recoveredID string) // TODO: change name, there is also a submitWorkflowSpec()
	AddCron(cron *core.Cron) (*core.Cron, error)
	RemoveGenerator(generatorID string) error
	GetCron(cronID string) (*core.Cron, error)
	GetCrons(colonyName string, count int) ([]*core.Cron, error)
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
}
