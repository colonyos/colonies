package server

import (
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
)

type controller interface {
	getCronPeriod() int
	getGeneratorPeriod() int
	getEtcdServer() *cluster.EtcdServer
	getEventHandler() *eventHandler
	getThisNode() cluster.Node
	subscribeProcesses(executorID string, subscription *subscription) error
	subscribeProcess(executorID string, subscription *subscription) error
	getColonies() ([]*core.Colony, error)
	getColony(colonyID string) (*core.Colony, error)
	addColony(colony *core.Colony) (*core.Colony, error)
	deleteColony(colonyID string) error
	renameColony(colonyID string, name string) error
	addExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error)
	addLog(processID string, colonyID string, executorID string, msg string) error
	getLogsByProcessID(processID string, limit int) ([]core.Log, error)
	getExecutor(executorID string) (*core.Executor, error)
	getExecutorByColonyID(colonyID string) ([]*core.Executor, error)
	approveExecutor(executorID string) error
	rejectExecutor(executorID string) error
	deleteExecutor(executorID string) error
	addProcessToDB(process *core.Process) (*core.Process, error)
	addProcess(process *core.Process) (*core.Process, error)
	addChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error)
	getProcess(processID string) (*core.Process, error)
	findProcessHistory(colonyID string, executorID string, seconds int, state int) ([]*core.Process, error)
	findWaitingProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	findRunningProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	findSuccessfulProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	findFailedProcesses(colonyID string, executorType string, count int) ([]*core.Process, error)
	updateProcessGraph(graph *core.ProcessGraph) error
	createProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}) (*core.ProcessGraph, error)
	submitWorkflowSpec(workflowSpec *core.WorkflowSpec) (*core.ProcessGraph, error)
	getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	findWaitingProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	findRunningProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	findSuccessfulProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	findFailedProcessGraphs(colonyID string, count int) ([]*core.ProcessGraph, error)
	deleteProcess(processID string) error
	deleteAllProcesses(colonyID string, state int) error
	deleteProcessGraph(processID string) error
	deleteAllProcessGraphs(colonyID string, state int) error
	setOutput(processID string, output []interface{}) error
	closeSuccessful(processID string, executorID string, output []interface{}) error
	notifyChildren(process *core.Process) error
	closeFailed(processID string, errs []string) error
	handleDefunctProcessgraph(processGraphID string, processID string, err error) error
	assign(executorID string, colonyID string) (*core.Process, error)
	unassignExecutor(processID string) error
	resetProcess(processID string) error
	getColonyStatistics(colonyID string) (*core.Statistics, error)
	getStatistics() (*core.Statistics, error)
	addAttribute(attribute *core.Attribute) (*core.Attribute, error)
	getAttribute(attributeID string) (*core.Attribute, error)
	addFunction(function *core.Function) (*core.Function, error)
	getFunctionsByExecutorID(executorID string) ([]*core.Function, error)
	getFunctionsByColonyID(colonyID string) ([]*core.Function, error)
	getFunctionByID(functionID string) (*core.Function, error)
	deleteFunction(functionID string) error
	addGenerator(generator *core.Generator) (*core.Generator, error)
	getGenerator(generatorID string) (*core.Generator, error)
	resolveGenerator(generatorName string) (*core.Generator, error)
	getGenerators(colonyID string, count int) ([]*core.Generator, error)
	packGenerator(generatorID string, colonyID, arg string) error
	generatorTriggerLoop()
	triggerGenerators()
	submitWorkflow(generator *core.Generator, counter int)
	addCron(cron *core.Cron) (*core.Cron, error)
	deleteGenerator(generatorID string) error
	getCron(cronID string) (*core.Cron, error)
	getCrons(colonyID string, count int) ([]*core.Cron, error)
	runCron(cronID string) (*core.Cron, error)
	deleteCron(cronID string) error
	calcNextRun(cron *core.Cron) time.Time
	startCron(cron *core.Cron)
	triggerCrons()
	cronTriggerLoop()
	resetDatabase() error
	stop()
	isLeader() bool
	tryBecomeLeader() bool
	timeoutLoop()
	blockingCmdQueueWorker()
	retentionWorker()
	cmdQueueWorker()
}
