package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupFakeServer() (*ColoniesServer, *controllerMock, *validatorMock, *dbMock, *gin.Context, *httptest.ResponseRecorder) {
	server := &ColoniesServer{}
	validatorMock := &validatorMock{}
	server.validator = validatorMock
	controllerMock := &controllerMock{}
	server.controller = controllerMock
	dbMock := &dbMock{}
	server.colonyDB = dbMock
	server.executorDB = dbMock
	server.processDB = dbMock
	server.userDB = dbMock
	server.functionDB = dbMock
	server.attributeDB = dbMock
	server.processGraphDB = dbMock
	server.generatorDB = dbMock
	server.cronDB = dbMock
	server.logDB = dbMock
	server.fileDB = dbMock
	server.snapshotDB = dbMock
	server.securityDB = dbMock
	ctx, w := getTestGinContext()

	return server, controllerMock, validatorMock, dbMock, ctx, w
}

func createFakeColoniesController() (*coloniesController, *dbMock) {
	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	dbMock := &dbMock{}
	return createColoniesController(dbMock, node, clusterConfig, "/tmp/colonies/etcd", GENERATOR_TRIGGER_PERIOD, CRON_TRIGGER_PERIOD, false, -1, 500), dbMock
}

// controllerMock
type controllerMock struct {
	returnError string
	returnValue string
}

func (v *controllerMock) getCronPeriod() int {
	return -1
}

func (v *controllerMock) getGeneratorPeriod() int {
	return -1
}

func (v *controllerMock) getEtcdServer() *cluster.EtcdServer {
	return nil
}

func (v *controllerMock) getEventHandler() *eventHandler {
	return nil
}

func (v *controllerMock) getThisNode() cluster.Node {
	return cluster.Node{}
}

func (v *controllerMock) subscribeProcesses(executorID string, subscription *subscription) error {
	return nil
}

func (v *controllerMock) subscribeProcess(executorID string, subscription *subscription) error {
	return nil
}

func (v *controllerMock) getColonies() ([]*core.Colony, error) {
	return nil, nil
}

func (v *controllerMock) getColony(colonyName string) (*core.Colony, error) {
	return nil, nil
}

func (v *controllerMock) addColony(colony *core.Colony) (*core.Colony, error) {
	return nil, nil
}

func (v *controllerMock) removeColony(colonyName string) error {
	return nil
}

func (v *controllerMock) addExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error) {
	return nil, nil
}

func (v *controllerMock) getExecutor(executorID string) (*core.Executor, error) {
	return nil, nil
}

func (v *controllerMock) getExecutorByColonyName(colonyName string) ([]*core.Executor, error) {
	return nil, nil
}

func (v *controllerMock) removeExecutor(executorID string) error {
	return nil
}

func (v *controllerMock) addProcessToDB(process *core.Process) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) addProcess(process *core.Process) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) addChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) getProcess(processID string) (*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) findProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (v *controllerMock) updateProcessGraph(graph *core.ProcessGraph) error {
	return nil
}

func (v *controllerMock) createProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) submitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) getProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) findFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *controllerMock) removeProcess(processID string) error {
	return nil
}

func (v *controllerMock) removeAllProcesses(colonyName string, state int) error {
	return nil
}

func (v *controllerMock) removeProcessGraph(processID string) error {
	return nil
}

func (v *controllerMock) removeAllProcessGraphs(colonyName string, state int) error {
	return nil
}

func (v *controllerMock) setOutput(processID string, output []interface{}) error {
	return nil
}

func (v *controllerMock) closeSuccessful(processID string, executorID string, output []interface{}) error {
	return nil
}

func (v *controllerMock) notifyChildren(process *core.Process) error {
	return nil
}

func (v *controllerMock) closeFailed(processID string, errs []string) error {
	return nil
}

func (v *controllerMock) handleDefunctProcessgraph(processGraphID string, processID string, err error) error {
	return nil
}

func (v *controllerMock) assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error) {
	result := &AssignResult{
		Process:       nil,
		IsPaused:      false,
		ResumeChannel: nil,
	}
	return result, nil
}

func (v *controllerMock) unassignExecutor(processID string) error {
	return nil
}

func (v *controllerMock) resetProcess(processID string) error {
	return nil
}

func (v *controllerMock) getColonyStatistics(colonyName string) (*core.Statistics, error) {
	return nil, nil
}

func (v *controllerMock) getStatistics() (*core.Statistics, error) {
	return nil, nil
}

func (v *controllerMock) addAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	return nil, nil
}

func (v *controllerMock) getAttribute(attributeID string) (*core.Attribute, error) {
	return nil, nil
}

func (v *controllerMock) addFunction(function *core.Function) (*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) getFunctionsByExecutorName(colonyName string, executorID string) ([]*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) getFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) getFunctionByID(functionID string) (*core.Function, error) {
	return nil, nil
}

func (v *controllerMock) removeFunction(functionID string) error {
	return nil
}

func (v *controllerMock) addGenerator(generator *core.Generator) (*core.Generator, error) {
	if v.returnError == "addGenerator" {
		return nil, errors.New("error")
	}

	return nil, nil
}

func (v *controllerMock) getGenerator(generatorID string) (*core.Generator, error) {
	if v.returnError == "getGenerator" {
		return nil, errors.New("error")
	}

	if v.returnValue == "getGenerator" {
		return &core.Generator{}, nil
	}

	return nil, nil
}

func (v *controllerMock) resolveGenerator(colonyName string, generatorName string) (*core.Generator, error) {
	return nil, nil
}

func (v *controllerMock) getGenerators(colonyName string, count int) ([]*core.Generator, error) {
	return nil, nil
}

func (v *controllerMock) packGenerator(generatorID string, colonyName, arg string) error {
	return nil
}

func (v *controllerMock) generatorTriggerLoop() {
}

func (v *controllerMock) triggerGenerators() {
}

func (v *controllerMock) submitWorkflow(generator *core.Generator, counter int, recoveredID string) {
}

func (v *controllerMock) addCron(cron *core.Cron) (*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) removeGenerator(generatorID string) error {
	return nil
}

func (v *controllerMock) getCron(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) getCrons(colonyName string, count int) ([]*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) runCron(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (v *controllerMock) removeCron(cronID string) error {
	return nil
}

func (v *controllerMock) calcNextRun(cron *core.Cron) time.Time {
	return time.Time{}
}

func (v *controllerMock) startCron(cron *core.Cron) {
}

func (v *controllerMock) triggerCrons() {
}

func (v *controllerMock) cronTriggerLoop() {
}

func (v *controllerMock) resetDatabase() error {
	return nil
}

func (v *controllerMock) stop() {
}

func (v *controllerMock) isLeader() bool {
	return false
}

func (v *controllerMock) tryBecomeLeader() bool {
	return false
}

func (v *controllerMock) timeoutLoop() {
}

func (v *controllerMock) blockingCmdQueueWorker() {
}

func (v *controllerMock) retentionWorker() {
}

func (v *controllerMock) cmdQueueWorker() {
}

func (v *controllerMock) pauseColonyAssignments(colonyName string) error {
	if v.returnError == "pauseColonyAssignments" {
		return errors.New("mock error")
	}
	return nil
}

func (v *controllerMock) resumeColonyAssignments(colonyName string) error {
	if v.returnError == "resumeColonyAssignments" {
		return errors.New("mock error")
	}
	return nil
}

func (v *controllerMock) areColonyAssignmentsPaused(colonyName string) (bool, error) {
	if v.returnError == "areColonyAssignmentsPaused" {
		return false, errors.New("mock error")
	}
	if v.returnValue == "paused" {
		return true, nil
	}
	return false, nil
}

// validatorMock
type validatorMock struct {
}

func (v *validatorMock) RequireServerOwner(recoveredID string, serverID string) error {
	return nil
}

func (v *validatorMock) RequireColonyOwner(recoveredID string, colonyName string) error {
	return nil
}

func (v *validatorMock) RequireMembership(recoveredID string, colonyName string, approved bool) error {
	return nil
}

type dbMock struct {
	returnError string
	returnValue string
}

func (db *dbMock) Close() {
}

func (db *dbMock) Initialize() error {
	return nil
}

func (db *dbMock) Drop() error {
	return nil
}

func (db *dbMock) AddUser(user *core.User) error {
	return nil
}

func (db *dbMock) GetUsersByColonyName(colonyName string) ([]*core.User, error) {
	return nil, nil
}

func (db *dbMock) GetUserByID(colonyName string, userID string) (*core.User, error) {
	return nil, nil
}

func (db *dbMock) GetUserByName(colonyName string, name string) (*core.User, error) {
	return nil, nil
}

func (db *dbMock) RemoveUserByID(colonyName string, userID string) error {
	return nil
}

func (db *dbMock) RemoveUserByName(colonyName string, name string) error {
	return nil
}

func (db *dbMock) RemoveUsersByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) AddColony(colony *core.Colony) error {
	if db.returnError == "AddColony" {
		return errors.New("error")
	}
	return nil
}

func (db *dbMock) GetColonies() ([]*core.Colony, error) {
	if db.returnError == "GetColonies" {
		return nil, errors.New("error")
	}

	return nil, nil
}

func (db *dbMock) GetColonyByID(id string) (*core.Colony, error) {
	if db.returnError == "GetColonyByID" {
		return nil, errors.New("error")
	}

	if db.returnValue == "GetColonyByID" {
		return &core.Colony{}, nil
	}

	return nil, nil
}

func (db *dbMock) GetColonyByName(name string) (*core.Colony, error) {
	if db.returnError == "GetColonyByName" {
		return nil, errors.New("error")
	}

	if db.returnValue == "GetColonyByName" {
		return &core.Colony{}, nil
	}

	return nil, nil
}

func (db *dbMock) RenameColony(id string, name string) error {
	if db.returnError == "RenameColony" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) RemoveColonyByName(colonyName string) error {
	if db.returnError == "RemoveColonyByName" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) CountColonies() (int, error) {
	return -1, nil
}

func (db *dbMock) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	return nil
}

func (db *dbMock) AddExecutor(executor *core.Executor) error {
	if db.returnError == "AddExecutor" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) GetExecutors() ([]*core.Executor, error) {
	return nil, nil
}

func (db *dbMock) GetExecutorByID(executorID string) (*core.Executor, error) {
	if db.returnError == "GetExecutorByID" {
		return nil, errors.New("error")
	}

	if db.returnValue == "GetExecutorByID" {
		return &core.Executor{}, nil
	}

	return nil, nil
}

func (db *dbMock) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	if db.returnError == "GetExecutorByColonyName" {
		return nil, errors.New("error")
	}

	return nil, nil
}

func (db *dbMock) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	if db.returnError == "GetExecutorByName" {
		return nil, errors.New("error")
	}

	if db.returnValue == "GetExecutorByName" {
		return &core.Executor{}, nil
	}

	return nil, nil
}

func (db *dbMock) ApproveExecutor(executor *core.Executor) error {
	if db.returnError == "ApproveExecutor" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) RejectExecutor(executor *core.Executor) error {
	if db.returnError == "RejectExecutor" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) MarkAlive(executor *core.Executor) error {
	return nil
}

func (db *dbMock) RemoveExecutorByName(colonyName string, executorName string) error {
	if db.returnError == "RemoveExecutorByName" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) RemoveExecutorsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) CountExecutors() (int, error) {
	return -1, nil
}

func (db *dbMock) CountExecutorsByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) AddFunction(function *core.Function) error {
	return nil
}

func (db *dbMock) GetFunctionByID(functionID string) (*core.Function, error) {
	return nil, nil
}

func (db *dbMock) GetFunctionsByExecutorName(colonyName string, executorID string) ([]*core.Function, error) {
	return nil, nil
}

func (db *dbMock) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	return nil, nil
}

func (db *dbMock) GetFunctionsByExecutorAndName(colonyName string, executorID string, name string) (*core.Function, error) {
	return nil, nil
}

func (db *dbMock) UpdateFunctionStats(colonyName string, executorID string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error {
	return nil
}

func (db *dbMock) RemoveFunctionByID(functionID string) error {
	return nil
}

func (db *dbMock) RemoveFunctionByName(colonyName string, executorName string, name string) error {
	return nil
}

func (db *dbMock) RemoveFunctionsByExecutorName(colonyName string, executorID string) error {
	return nil
}

func (db *dbMock) RemoveFunctionsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveFunctions() error {
	return nil
}

func (db *dbMock) AddProcess(process *core.Process) error {
	if db.returnError == "AddProcess" {
		return errors.New("error")
	}

	return nil
}

func (db *dbMock) GetProcesses() ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) GetProcessByID(processID string) (*core.Process, error) {
	if db.returnError == "GetProcessByID" {
		return nil, errors.New("error")
	}
	return nil, nil
}

func (db *dbMock) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindAllRunningProcesses() ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindAllWaitingProcesses() ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindCandidates(colonyName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) FindCandidatesByName(colonyName string, executorName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) {
	return nil, nil
}

func (db *dbMock) RemoveProcessByID(processID string) error {
	return nil
}

func (db *dbMock) RemoveAllProcesses() error {
	return nil
}

func (db *dbMock) RemoveAllWaitingProcessesByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllRunningProcessesByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllFailedProcessesByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllProcessesByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllProcessesByProcessGraphID(processGraphID string) error {
	return nil
}

func (db *dbMock) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) ResetProcess(process *core.Process) error {
	return nil
}

func (db *dbMock) SetInput(processID string, output []interface{}) error {
	return nil
}

func (db *dbMock) SetOutput(processID string, output []interface{}) error {
	return nil
}

func (db *dbMock) SetErrors(processID string, errs []string) error {
	return nil
}

func (db *dbMock) SetProcessState(processID string, state int) error {
	return nil
}

func (db *dbMock) SetParents(processID string, parents []string) error {
	return nil
}

func (db *dbMock) SetChildren(processID string, children []string) error {
	return nil
}

func (db *dbMock) SetWaitForParents(processID string, waitingForParent bool) error {
	return nil
}

func (db *dbMock) Assign(executorID string, process *core.Process) error {
	return nil
}

func (db *dbMock) Unassign(process *core.Process) error {
	return nil
}

func (db *dbMock) MarkSuccessful(processID string) (float64, float64, error) {
	return -1.0, -1.0, nil
}

func (db *dbMock) MarkFailed(processID string, errs []string) error {
	return nil
}

func (db *dbMock) CountProcesses() (int, error) {
	return -1, nil
}

func (db *dbMock) CountWaitingProcesses() (int, error) {
	return -1, nil
}

func (db *dbMock) CountRunningProcesses() (int, error) {
	return -1, nil
}

func (db *dbMock) CountSuccessfulProcesses() (int, error) {
	return -1, nil
}

func (db *dbMock) CountFailedProcesses() (int, error) {
	return -1, nil
}

func (db *dbMock) CountWaitingProcessesByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) CountRunningProcessesByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) CountFailedProcessesByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) AddAttribute(attribute core.Attribute) error {
	return nil
}

func (db *dbMock) AddAttributes(attribute []core.Attribute) error {
	return nil
}

func (db *dbMock) GetAttributeByID(attributeID string) (core.Attribute, error) {
	return core.Attribute{}, nil
}

func (db *dbMock) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) {
	return nil, nil
}

func (db *dbMock) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) {
	return core.Attribute{}, nil
}

func (db *dbMock) GetAttributes(targetID string) ([]core.Attribute, error) {
	return nil, nil
}

func (db *dbMock) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) {
	return nil, nil
}

func (db *dbMock) UpdateAttribute(attribute core.Attribute) error {
	return nil
}

func (db *dbMock) RemoveAttributeByID(attributeID string) error {
	return nil
}

func (db *dbMock) RemoveAllAttributesByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error {
	return nil
}

func (db *dbMock) RemoveAllAttributesByProcessGraphID(processGraphID string) error {
	return nil
}

func (db *dbMock) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error {
	return nil
}

func (db *dbMock) RemoveAttributesByTargetID(targetID string, attributeType int) error {
	return nil
}

func (db *dbMock) RemoveAllAttributesByTargetID(targetID string) error {
	return nil
}

func (db *dbMock) RemoveAllAttributes() error {
	return nil
}

func (db *dbMock) AddProcessGraph(processGraph *core.ProcessGraph) error {
	return nil
}

func (db *dbMock) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (db *dbMock) SetProcessGraphState(processGraphID string, state int) error {
	return nil
}

func (db *dbMock) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (db *dbMock) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (db *dbMock) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (db *dbMock) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (db *dbMock) RemoveProcessGraphByID(processGraphID string) error {
	return nil
}

func (db *dbMock) RemoveAllProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) CountWaitingProcessGraphs() (int, error) {
	return -1, nil
}

func (db *dbMock) CountRunningProcessGraphs() (int, error) {
	return -1, nil
}

func (db *dbMock) CountSuccessfulProcessGraphs() (int, error) {
	return -1, nil
}

func (db *dbMock) CountFailedProcessGraphs() (int, error) {
	return -1, nil
}

func (db *dbMock) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) {
	return -1, nil
}

func (db *dbMock) AddGenerator(generator *core.Generator) error {
	return nil
}

func (db *dbMock) SetGeneratorLastRun(generatorID string) error {
	return nil
}

func (db *dbMock) SetGeneratorFirstPack(generatorID string) error {
	return nil
}

func (db *dbMock) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	return nil, nil
}

func (db *dbMock) GetGeneratorByName(colonyName string, name string) (*core.Generator, error) {
	return nil, nil
}

func (db *dbMock) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) {
	return nil, nil
}

func (db *dbMock) FindAllGenerators() ([]*core.Generator, error) {
	return nil, nil
}

func (db *dbMock) RemoveGeneratorByID(generatorID string) error {
	return nil
}

func (db *dbMock) RemoveAllGeneratorsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) AddGeneratorArg(generatorArg *core.GeneratorArg) error {
	return nil
}

func (db *dbMock) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) {
	return nil, nil
}

func (db *dbMock) CountGeneratorArgs(generatorID string) (int, error) {
	if db.returnError == "CountGeneratorArgs" {
		return -1, errors.New("error")

	}
	return -1, nil
}

func (db *dbMock) RemoveGeneratorArgByID(generatorArgsID string) error {
	return nil
}

func (db *dbMock) RemoveAllGeneratorArgsByGeneratorID(generatorID string) error {
	return nil
}

func (db *dbMock) RemoveAllGeneratorArgsByColonyName(generatorID string) error {
	return nil
}

func (db *dbMock) AddCron(cron *core.Cron) error {
	return nil
}

func (db *dbMock) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, lastProcessGraphID string) error {
	return nil
}

func (db *dbMock) GetCronByID(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (db *dbMock) GetCronByName(colonyName string, cronName string) (*core.Cron, error) {
	return nil, nil
}

func (db *dbMock) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error) {
	return nil, nil
}

func (db *dbMock) FindAllCrons() ([]*core.Cron, error) {
	return nil, nil
}

func (db *dbMock) RemoveCronByID(cronID string) error {
	return nil
}

func (db *dbMock) RemoveAllCronsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) Lock(timeout int) error {

	return nil
}

func (db *dbMock) Unlock() error {
	return nil
}

func (db *dbMock) ApplyRetentionPolicy(retentionPeriod int64) error {
	return nil
}

func (db *dbMock) AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error {
	return nil
}

func (db *dbMock) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) {
	return []*core.Log{}, nil
}

func (db *dbMock) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) {
	return []*core.Log{}, nil
}

func (db *dbMock) GetLogsByExecutor(processID string, limit int) ([]*core.Log, error) {
	return []*core.Log{}, nil
}

func (db *dbMock) GetLogsByExecutorSince(processID string, limit int, since int64) ([]*core.Log, error) {
	return []*core.Log{}, nil
}

func (db *dbMock) RemoveLogsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) CountLogs(colonyName string) (int, error) {
	return 0, nil
}

func (db *dbMock) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) {
	return nil, nil
}

func (db *dbMock) AddFile(file *core.File) error {
	return nil
}

func (db *dbMock) GetFileByID(colonyName string, fileID string) (*core.File, error) {
	return nil, nil
}

func (db *dbMock) GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	return nil, nil
}

func (db *dbMock) GetFileByName(colonyName string, label string, name string) ([]*core.File, error) {
	return nil, nil
}

func (db *dbMock) GetFilenamesByLabel(colonyName string, label string) ([]string, error) {
	return nil, nil
}

func (db *dbMock) GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error) {
	return nil, nil
}

func (db *dbMock) RemoveFileByID(colonyName string, fileID string) error {
	return nil
}

func (db *dbMock) RemoveFileByName(colonyName string, label string, name string) error {
	return nil
}

func (db *dbMock) GetFileLabels(colonyName string) ([]*core.Label, error) {
	return nil, nil
}

func (db *dbMock) GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error) {
	return nil, nil
}

func (db *dbMock) CountFiles(colonyName string) (int, error) {
	return 0, nil
}

func (db *dbMock) CountFilesWithLabel(colonyName string, label string) (int, error) {
	return 0, nil
}

func (db *dbMock) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error) {
	return nil, nil
}

func (db *dbMock) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error) {
	return nil, nil
}

func (db *dbMock) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error) {
	return nil, nil
}

func (db *dbMock) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error) {
	return nil, nil
}

func (db *dbMock) RemoveSnapshotByID(colonyName string, snapshotID string) error {
	return nil
}

func (db *dbMock) RemoveSnapshotByName(colonyName string, name string) error {
	return nil
}

func (db *dbMock) RemoveSnapshotsByColonyName(colonyName string) error {
	return nil
}

func (db *dbMock) SetServerID(oldServerID, newServerID string) error {
	return nil
}

func (db *dbMock) GetServerID() (string, error) {
	return "", nil
}

func (db *dbMock) ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error {
	return nil
}

func (db *dbMock) ChangeUserID(colonyName string, oldUserID, newUserID string) error {
	return nil
}

func (db *dbMock) ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error {
	return nil
}

// gin mockups
func getTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = &http.Request{
		Header: make(http.Header),
	}

	return ctx, w
}

func assertRPCError(t *testing.T, body string) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(body)
	assert.Nil(t, err)
	assert.True(t, rpcReplyMsg.Error)
}
