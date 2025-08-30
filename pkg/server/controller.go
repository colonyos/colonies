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
	getColony(colonyName string) (*core.Colony, error)
	addColony(colony *core.Colony) (*core.Colony, error)
	removeColony(colonyName string) error
	addExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error)
	getExecutor(executorID string) (*core.Executor, error)
	getExecutorByColonyName(colonyName string) ([]*core.Executor, error)
	addProcessToDB(process *core.Process) (*core.Process, error)
	addProcess(process *core.Process) (*core.Process, error)
	addChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error)
	getProcess(processID string) (*core.Process, error)
	findProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error)
	updateProcessGraph(graph *core.ProcessGraph) error
	createProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error)
	submitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error)
	getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error)
	findWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	findRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	findSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	findFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error)
	removeProcess(processID string) error
	removeAllProcesses(colonyName string, state int) error
	removeProcessGraph(processID string) error
	removeAllProcessGraphs(colonyName string, state int) error
	setOutput(processID string, output []interface{}) error
	closeSuccessful(processID string, executorID string, output []interface{}) error
	notifyChildren(process *core.Process) error
	closeFailed(processID string, errs []string) error
	handleDefunctProcessgraph(processGraphID string, processID string, err error) error
	assign(executorID string, colonyName string, cpu int64, memory int64) (*core.Process, error)
	unassignExecutor(processID string) error
	resetProcess(processID string) error
	getColonyStatistics(colonyName string) (*core.Statistics, error)
	getStatistics() (*core.Statistics, error)
	addAttribute(attribute *core.Attribute) (*core.Attribute, error)
	getAttribute(attributeID string) (*core.Attribute, error)
	addFunction(function *core.Function) (*core.Function, error)
	getFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error)
	getFunctionsByColonyName(colonyName string) ([]*core.Function, error)
	getFunctionByID(functionID string) (*core.Function, error)
	removeFunction(functionID string) error
	addGenerator(generator *core.Generator) (*core.Generator, error)
	getGenerator(generatorID string) (*core.Generator, error)
	resolveGenerator(colonyName string, generatorName string) (*core.Generator, error)
	getGenerators(colonyName string, count int) ([]*core.Generator, error)
	packGenerator(generatorID string, colonyName, arg string) error
	generatorTriggerLoop()
	triggerGenerators()
	submitWorkflow(generator *core.Generator, counter int, recoveredID string) // TODO: change name, there is also a submitWorkflowSpec()
	addCron(cron *core.Cron) (*core.Cron, error)
	removeGenerator(generatorID string) error
	getCron(cronID string) (*core.Cron, error)
	getCrons(colonyName string, count int) ([]*core.Cron, error)
	runCron(cronID string) (*core.Cron, error)
	removeCron(cronID string) error
	calcNextRun(cron *core.Cron) time.Time
	startCron(cron *core.Cron)
	triggerCrons()
	cronTriggerLoop()
	resetDatabase() error
	pauseColonyAssignments(colonyName string) error
	resumeColonyAssignments(colonyName string) error
	areColonyAssignmentsPaused(colonyName string) (bool, error)
	stop()
	isLeader() bool
	tryBecomeLeader() bool
	timeoutLoop()
	blockingCmdQueueWorker()
	retentionWorker()
	cmdQueueWorker()
}
