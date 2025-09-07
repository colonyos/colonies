package controllers

import (
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/backends"
)

// ControllerMock implements the Controller interface for testing
type ControllerMock struct {
	ReturnError string
	ReturnValue string
}

func (v *ControllerMock) GetCronPeriod() int {
	return -1
}

func (v *ControllerMock) GetGeneratorPeriod() int {
	return -1
}

func (v *ControllerMock) GetEtcdServer() *cluster.EtcdServer {
	return nil
}

func (v *ControllerMock) GetEventHandler() backends.RealtimeEventHandler {
	return nil
}

func (v *ControllerMock) GetThisNode() cluster.Node {
	return cluster.Node{}
}

func (v *ControllerMock) SubscribeProcesses(executorID string, subscription *backends.RealtimeSubscription) error {
	return nil
}

func (v *ControllerMock) SubscribeProcess(executorID string, subscription *backends.RealtimeSubscription) error {
	return nil
}

func (v *ControllerMock) GetColonies() ([]*core.Colony, error) {
	return nil, nil
}

func (v *ControllerMock) GetColony(colonyName string) (*core.Colony, error) {
	return nil, nil
}

func (v *ControllerMock) AddColony(colony *core.Colony) (*core.Colony, error) {
	return nil, nil
}

func (v *ControllerMock) RemoveColony(colonyName string) error {
	return nil
}

func (v *ControllerMock) AddExecutor(executor *core.Executor, allowExecutorReregister bool) (*core.Executor, error) {
	return nil, nil
}

func (v *ControllerMock) GetExecutor(executorID string) (*core.Executor, error) {
	return nil, nil
}

func (v *ControllerMock) GetExecutorByColonyName(colonyName string) ([]*core.Executor, error) {
	return nil, nil
}

func (v *ControllerMock) AddProcessToDB(process *core.Process) (*core.Process, error) {
	return nil, nil
}

func (v *ControllerMock) AddProcess(process *core.Process) (*core.Process, error) {
	return nil, nil
}

func (v *ControllerMock) AddChild(processGraphID string, parentProcessID string, childProcessID string, process *core.Process, executorID string, insert bool) (*core.Process, error) {
	return nil, nil
}

func (v *ControllerMock) GetProcess(processID string) (*core.Process, error) {
	return nil, nil
}

func (v *ControllerMock) FindProcessHistory(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) {
	return nil, nil
}

func (v *ControllerMock) UpdateProcessGraph(graph *core.ProcessGraph) error {
	return nil
}

func (v *ControllerMock) CreateProcessGraph(workflowSpec *core.WorkflowSpec, args []interface{}, kwargs map[string]interface{}, rootInput []interface{}, recoveredID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) SubmitWorkflowSpec(workflowSpec *core.WorkflowSpec, recoveredID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) {
	return nil, nil
}

func (v *ControllerMock) RemoveProcess(processID string) error {
	return nil
}

func (v *ControllerMock) RemoveAllProcesses(colonyName string, state int) error {
	return nil
}

func (v *ControllerMock) RemoveProcessGraph(processID string) error {
	return nil
}

func (v *ControllerMock) RemoveAllProcessGraphs(colonyName string, state int) error {
	return nil
}

func (v *ControllerMock) SetOutput(processID string, output []interface{}) error {
	return nil
}

func (v *ControllerMock) CloseSuccessful(processID string, executorID string, output []interface{}) error {
	return nil
}

func (v *ControllerMock) NotifyChildren(process *core.Process) error {
	return nil
}

func (v *ControllerMock) CloseFailed(processID string, errs []string) error {
	return nil
}

func (v *ControllerMock) HandleDefunctProcessgraph(processGraphID string, processID string, err error) error {
	return nil
}

func (v *ControllerMock) Assign(executorID string, colonyName string, cpu int64, memory int64) (*AssignResult, error) {
	return nil, nil
}

func (v *ControllerMock) UnassignExecutor(processID string) error {
	return nil
}

func (v *ControllerMock) ResetProcess(processID string) error {
	return nil
}

func (v *ControllerMock) GetColonyStatistics(colonyName string) (*core.Statistics, error) {
	return nil, nil
}

func (v *ControllerMock) GetStatistics() (*core.Statistics, error) {
	return nil, nil
}

func (v *ControllerMock) AddAttribute(attribute *core.Attribute) (*core.Attribute, error) {
	return nil, nil
}

func (v *ControllerMock) GetAttribute(attributeID string) (*core.Attribute, error) {
	return nil, nil
}

func (v *ControllerMock) AddFunction(function *core.Function) (*core.Function, error) {
	return nil, nil
}

func (v *ControllerMock) GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error) {
	return nil, nil
}

func (v *ControllerMock) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) {
	return nil, nil
}

func (v *ControllerMock) GetFunctionByID(functionID string) (*core.Function, error) {
	return nil, nil
}

func (v *ControllerMock) RemoveFunction(functionID string) error {
	return nil
}

func (v *ControllerMock) AddGenerator(generator *core.Generator) (*core.Generator, error) {
	return nil, nil
}

func (v *ControllerMock) GetGenerator(generatorID string) (*core.Generator, error) {
	return nil, nil
}

func (v *ControllerMock) ResolveGenerator(colonyName string, generatorName string) (*core.Generator, error) {
	return nil, nil
}

func (v *ControllerMock) GetGenerators(colonyName string, count int) ([]*core.Generator, error) {
	return nil, nil
}

func (v *ControllerMock) PackGenerator(generatorID string, colonyName, arg string) error {
	return nil
}

func (v *ControllerMock) GeneratorTriggerLoop() {
}

func (v *ControllerMock) TriggerGenerators() {
}

func (v *ControllerMock) SubmitWorkflow(generator *core.Generator, counter int, recoveredID string) {
}

func (v *ControllerMock) AddCron(cron *core.Cron) (*core.Cron, error) {
	return nil, nil
}

func (v *ControllerMock) RemoveGenerator(generatorID string) error {
	return nil
}

func (v *ControllerMock) GetCron(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (v *ControllerMock) GetCrons(colonyName string, count int) ([]*core.Cron, error) {
	return nil, nil
}

func (v *ControllerMock) RunCron(cronID string) (*core.Cron, error) {
	return nil, nil
}

func (v *ControllerMock) RemoveCron(cronID string) error {
	return nil
}

func (v *ControllerMock) CalcNextRun(cron *core.Cron) time.Time {
	return time.Time{}
}

func (v *ControllerMock) StartCron(cron *core.Cron) {
}

func (v *ControllerMock) TriggerCrons() {
}

func (v *ControllerMock) CronTriggerLoop() {
}

func (v *ControllerMock) ResetDatabase() error {
	return nil
}

func (v *ControllerMock) PauseColonyAssignments(colonyName string) error {
	return nil
}

func (v *ControllerMock) ResumeColonyAssignments(colonyName string) error {
	return nil
}

func (v *ControllerMock) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	return false, nil
}

func (v *ControllerMock) Stop() {
}

func (v *ControllerMock) IsLeader() bool {
	return false
}

func (v *ControllerMock) TryBecomeLeader() bool {
	return false
}

func (v *ControllerMock) TimeoutLoop() {
}

func (v *ControllerMock) BlockingCmdQueueWorker() {
}

func (v *ControllerMock) RetentionWorker() {
}

func (v *ControllerMock) CmdQueueWorker() {
}

// DatabaseMock implements database interfaces for testing
type DatabaseMock struct {
	ReturnError string
	ReturnValue string
}

// Implement all database interfaces as no-ops for testing
// ColonyDatabase interface
func (db *DatabaseMock) AddColony(colony *core.Colony) error {
	if db.ReturnError == "AddColony" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) GetColonies() ([]*core.Colony, error) {
	if db.ReturnError == "GetColonies" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) GetColonyByID(id string) (*core.Colony, error) {
	if db.ReturnError == "GetColonyByID" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) GetColonyByName(name string) (*core.Colony, error) {
	if db.ReturnError == "GetColonyByName" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) RenameColony(colonyName string, newColonyName string) error {
	if db.ReturnError == "RenameColony" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) RemoveColonyByName(colonyName string) error {
	if db.ReturnError == "RemoveColonyByName" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) CountColonies() (int, error) {
	if db.ReturnError == "CountColonies" { return 0, errors.New("mock error") }
	return 0, nil
}

// ExecutorDatabase interface  
func (db *DatabaseMock) AddExecutor(executor *core.Executor) error {
	if db.ReturnError == "AddExecutor" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) SetAllocations(colonyName string, executorName string, allocations core.Allocations) error {
	if db.ReturnError == "SetAllocations" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) GetExecutors() ([]*core.Executor, error) {
	if db.ReturnError == "GetExecutors" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) GetExecutorByID(executorID string) (*core.Executor, error) {
	if db.ReturnError == "GetExecutorByID" { return nil, errors.New("mock error") }
	// Return a dummy executor when no error is set
	return &core.Executor{ID: executorID, ColonyName: "test-colony"}, nil
}
func (db *DatabaseMock) GetExecutorsByColonyName(colonyName string) ([]*core.Executor, error) {
	if db.ReturnError == "GetExecutorsByColonyName" || db.ReturnError == "GetExecutorByColonyName" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) GetExecutorByName(colonyName string, executorName string) (*core.Executor, error) {
	if db.ReturnError == "GetExecutorByName" { return nil, errors.New("mock error") }
	if db.ReturnValue == "GetExecutorByName" { return &core.Executor{}, nil }
	return nil, nil
}
func (db *DatabaseMock) ApproveExecutor(executor *core.Executor) error { return nil }
func (db *DatabaseMock) RejectExecutor(executor *core.Executor) error { return nil }
func (db *DatabaseMock) MarkAlive(executor *core.Executor) error { return nil }
func (db *DatabaseMock) RemoveExecutorByName(colonyName string, executorName string) error { return nil }
func (db *DatabaseMock) RemoveExecutorsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) CountExecutors() (int, error) { return 0, nil }
func (db *DatabaseMock) CountExecutorsByColonyName(colonyName string) (int, error) { return 0, nil }

// ProcessDatabase interface
func (db *DatabaseMock) AddProcess(process *core.Process) error {
	if db.ReturnError == "AddProcess" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) GetProcesses() ([]*core.Process, error) {
	if db.ReturnError == "GetProcesses" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) GetProcessByID(processID string) (*core.Process, error) {
	if db.ReturnError == "GetProcessByID" { return nil, errors.New("mock error") }
	return nil, nil
}
func (db *DatabaseMock) SetProcessState(processID string, state int) error {
	if db.ReturnError == "SetProcessState" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) SetWaitForParents(processID string, waitForParent bool) error {
	if db.ReturnError == "SetWaitForParents" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) MarkSuccessful(processID string) (float64, float64, error) {
	if db.ReturnError == "MarkSuccessful" { return 0, 0, errors.New("mock error") }
	return 1.0, 1.0, nil
}
func (db *DatabaseMock) FindProcessesByColonyName(colonyName string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindProcessesByExecutorID(colonyName string, executorID string, seconds int, state int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindWaitingProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindRunningProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindSuccessfulProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindFailedProcesses(colonyName string, executorType string, label string, initiator string, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindAllRunningProcesses() ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindAllWaitingProcesses() ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindCandidates(colonyName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) FindCandidatesByName(colonyName string, executorName string, executorType string, cpu int64, memory int64, storage int64, nodes int, processes int, processesPerNode int, count int) ([]*core.Process, error) { return nil, nil }
func (db *DatabaseMock) RemoveProcessByID(processID string) error { return nil }
func (db *DatabaseMock) RemoveAllProcesses() error { return nil }
func (db *DatabaseMock) RemoveAllWaitingProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllRunningProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllSuccessfulProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllFailedProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllProcessesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllProcessesByProcessGraphID(processGraphID string) error { return nil }
func (db *DatabaseMock) RemoveAllProcessesInProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) ResetProcess(process *core.Process) error { return nil }
func (db *DatabaseMock) SetInput(processID string, input []interface{}) error { return nil }
func (db *DatabaseMock) SetOutput(processID string, output []interface{}) error { return nil }
func (db *DatabaseMock) SetErrors(processID string, errs []string) error { return nil }
func (db *DatabaseMock) SetParents(processID string, parents []string) error { return nil }
func (db *DatabaseMock) SetChildren(processID string, children []string) error { return nil }
func (db *DatabaseMock) Assign(executorID string, process *core.Process) error { return nil }
func (db *DatabaseMock) Unassign(process *core.Process) error { return nil }
func (db *DatabaseMock) MarkFailed(processID string, errs []string) error { return nil }
func (db *DatabaseMock) CountProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountWaitingProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountRunningProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountSuccessfulProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountFailedProcesses() (int, error) { return 0, nil }
func (db *DatabaseMock) CountWaitingProcessesByColonyName(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountRunningProcessesByColonyName(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountSuccessfulProcessesByColonyName(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountFailedProcessesByColonyName(colonyName string) (int, error) { return 0, nil }

// UserDatabase interface
func (db *DatabaseMock) AddUser(user *core.User) error { return nil }
func (db *DatabaseMock) GetUserByName(colonyName string, name string) (*core.User, error) { return nil, nil }
func (db *DatabaseMock) GetUserByID(colonyName string, userID string) (*core.User, error) { return nil, nil }
func (db *DatabaseMock) GetUsersByColonyName(colonyName string) ([]*core.User, error) { return nil, nil }
func (db *DatabaseMock) RemoveUserByName(colonyName string, name string) error { return nil }
func (db *DatabaseMock) RemoveUserByID(colonyName string, userID string) error { return nil }
func (db *DatabaseMock) RemoveUsersByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) CountUsers() (int, error) { return 0, nil }

// AttributeDatabase interface
func (db *DatabaseMock) AddAttribute(attribute core.Attribute) error { return nil }
func (db *DatabaseMock) AddAttributes(attributes []core.Attribute) error { return nil }
func (db *DatabaseMock) GetAttributeByID(attributeID string) (core.Attribute, error) { return core.Attribute{}, nil }
func (db *DatabaseMock) GetAttributesByColonyName(colonyName string) ([]core.Attribute, error) { return nil, nil }
func (db *DatabaseMock) GetAttribute(targetID string, key string, attributeType int) (core.Attribute, error) { return core.Attribute{}, nil }
func (db *DatabaseMock) GetAttributes(targetID string) ([]core.Attribute, error) { return nil, nil }
func (db *DatabaseMock) GetAttributesByType(targetID string, attributeType int) ([]core.Attribute, error) { return nil, nil }
func (db *DatabaseMock) UpdateAttribute(attribute core.Attribute) error { return nil }
func (db *DatabaseMock) RemoveAttributeByID(attributeID string) error { return nil }
func (db *DatabaseMock) RemoveAllAttributesByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllAttributesByColonyNameWithState(colonyName string, state int) error { return nil }
func (db *DatabaseMock) RemoveAllAttributesByProcessGraphID(processGraphID string) error { return nil }
func (db *DatabaseMock) RemoveAllAttributesInProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllAttributesInProcessGraphsByColonyNameWithState(colonyName string, state int) error { return nil }
func (db *DatabaseMock) RemoveAttributesByTargetID(targetID string, attributeType int) error { return nil }
func (db *DatabaseMock) RemoveAllAttributesByTargetID(targetID string) error { return nil }
func (db *DatabaseMock) RemoveAllAttributes() error { return nil }

// FunctionDatabase interface
func (db *DatabaseMock) AddFunction(function *core.Function) error { return nil }
func (db *DatabaseMock) GetFunctionByID(functionID string) (*core.Function, error) { return nil, nil }
func (db *DatabaseMock) GetFunctionsByExecutorName(colonyName string, executorName string) ([]*core.Function, error) { return nil, nil }
func (db *DatabaseMock) GetFunctionsByColonyName(colonyName string) ([]*core.Function, error) { return nil, nil }
func (db *DatabaseMock) RemoveFunctionByID(functionID string) error { return nil }
func (db *DatabaseMock) RemoveFunctionsByExecutorName(colonyName string, executorName string) error { return nil }
func (db *DatabaseMock) RemoveFunctionsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) GetFunctionsByExecutorAndName(colonyName string, executorName string, name string) (*core.Function, error) { return nil, nil }
func (db *DatabaseMock) UpdateFunctionStats(colonyName string, executorName string, name string, counter int, minWaitTime float64, maxWaitTime float64, minExecTime float64, maxExecTime float64, avgWaitTime float64, avgExecTime float64) error { return nil }
func (db *DatabaseMock) RemoveFunctionByName(colonyName string, executorName string, name string) error { return nil }
func (db *DatabaseMock) RemoveFunctions() error { return nil }
func (db *DatabaseMock) CountFunctions() (int, error) { return 0, nil }

// ProcessGraphDatabase interface
func (db *DatabaseMock) AddProcessGraph(processGraph *core.ProcessGraph) error { return nil }
func (db *DatabaseMock) GetProcessGraphByID(processGraphID string) (*core.ProcessGraph, error) { return nil, nil }
func (db *DatabaseMock) SetProcessGraphState(processGraphID string, state int) error {
	if db.ReturnError == "SetProcessGraphState" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) FindWaitingProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (db *DatabaseMock) FindRunningProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (db *DatabaseMock) FindSuccessfulProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (db *DatabaseMock) FindFailedProcessGraphs(colonyName string, count int) ([]*core.ProcessGraph, error) { return nil, nil }
func (db *DatabaseMock) RemoveProcessGraphByID(processGraphID string) error { return nil }
func (db *DatabaseMock) RemoveAllProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllWaitingProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllRunningProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllSuccessfulProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllFailedProcessGraphsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) CountWaitingProcessGraphs() (int, error) { return 0, nil }
func (db *DatabaseMock) CountRunningProcessGraphs() (int, error) { return 0, nil }
func (db *DatabaseMock) CountSuccessfulProcessGraphs() (int, error) { return 0, nil }
func (db *DatabaseMock) CountFailedProcessGraphs() (int, error) { return 0, nil }
func (db *DatabaseMock) CountWaitingProcessGraphsByColonyName(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountRunningProcessGraphsByColonyName(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountSuccessfulProcessGraphsByColonyName(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountFailedProcessGraphsByColonyName(colonyName string) (int, error) { return 0, nil }

// GeneratorDatabase interface
func (db *DatabaseMock) AddGenerator(generator *core.Generator) error {
	if db.ReturnError == "AddGenerator" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) GetGeneratorByID(generatorID string) (*core.Generator, error) {
	if db.ReturnError == "GetGeneratorByID" { return nil, errors.New("mock error") }
	return &core.Generator{ID: generatorID, ColonyName: "test-colony"}, nil
}
func (db *DatabaseMock) GetGeneratorByName(colonyName string, generatorName string) (*core.Generator, error) {
	if db.ReturnError == "GetGeneratorByName" { return nil, errors.New("mock error") }
	return &core.Generator{ID: generatorName, ColonyName: colonyName}, nil
}
func (db *DatabaseMock) FindAllGenerators() ([]*core.Generator, error) { return nil, nil }
func (db *DatabaseMock) FindGeneratorsByColonyName(colonyName string, count int) ([]*core.Generator, error) { return nil, nil }
func (db *DatabaseMock) RemoveGeneratorByID(generatorID string) error { return nil }
func (db *DatabaseMock) AddGeneratorArg(generatorArg *core.GeneratorArg) error { return nil }
func (db *DatabaseMock) GetGeneratorArgs(generatorID string, count int) ([]*core.GeneratorArg, error) { return nil, nil }
func (db *DatabaseMock) CountGeneratorArgs(generatorID string) (int, error) { return 0, nil }
func (db *DatabaseMock) RemoveGeneratorArgByID(generatorArgID string) error { return nil }
func (db *DatabaseMock) SetGeneratorLastRun(generatorID string) error { return nil }
func (db *DatabaseMock) SetGeneratorFirstPack(generatorID string) error { return nil }
func (db *DatabaseMock) RemoveAllGeneratorsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) RemoveAllGeneratorArgsByGeneratorID(generatorID string) error { return nil }
func (db *DatabaseMock) RemoveAllGeneratorArgsByColonyName(colonyName string) error { return nil }

// CronDatabase interface
func (db *DatabaseMock) AddCron(cron *core.Cron) error {
	if db.ReturnError == "AddCron" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) GetCronByID(cronID string) (*core.Cron, error) {
	if db.ReturnError == "GetCronByID" { return nil, errors.New("mock error") }
	return &core.Cron{ID: cronID, ColonyName: "test-colony"}, nil
}
func (db *DatabaseMock) FindAllCrons() ([]*core.Cron, error) { return nil, nil }
func (db *DatabaseMock) FindCronsByColonyName(colonyName string, count int) ([]*core.Cron, error) { return nil, nil }
func (db *DatabaseMock) RemoveCronByID(cronID string) error {
	if db.ReturnError == "RemoveCronByID" { return errors.New("mock error") }
	return nil
}
func (db *DatabaseMock) UpdateCron(cronID string, nextRun time.Time, lastRun time.Time, processGraphID string) error { return nil }
func (db *DatabaseMock) GetCronByName(colonyName string, cronName string) (*core.Cron, error) { return nil, nil }
func (db *DatabaseMock) RemoveAllCronsByColonyName(colonyName string) error { return nil }

// LogDatabase interface  
func (db *DatabaseMock) AddLog(processID string, colonyName string, executorName string, timestamp int64, msg string) error { return nil }
func (db *DatabaseMock) GetLogsByProcessID(processID string, limit int) ([]*core.Log, error) { return nil, nil }
func (db *DatabaseMock) GetLogsByProcessIDSince(processID string, limit int, since int64) ([]*core.Log, error) { return nil, nil }
func (db *DatabaseMock) GetLogsByExecutor(executorName string, limit int) ([]*core.Log, error) { return nil, nil }
func (db *DatabaseMock) GetLogsByExecutorSince(executorName string, limit int, since int64) ([]*core.Log, error) { return nil, nil }
func (db *DatabaseMock) RemoveLogsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) CountLogs(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) SearchLogs(colonyName string, text string, days int, count int) ([]*core.Log, error) { return nil, nil }

// FileDatabase interface
func (db *DatabaseMock) AddFile(file *core.File) error { return nil }
func (db *DatabaseMock) GetFileByID(colonyName string, fileID string) (*core.File, error) { return nil, nil }
func (db *DatabaseMock) GetLatestFileByName(colonyName string, label string, name string) ([]*core.File, error) { return nil, nil }
func (db *DatabaseMock) GetFileByName(colonyName string, label string, name string) ([]*core.File, error) { return nil, nil }
func (db *DatabaseMock) GetFilenamesByLabel(colonyName string, label string) ([]string, error) { return nil, nil }
func (db *DatabaseMock) GetFileDataByLabel(colonyName string, label string) ([]*core.FileData, error) { return nil, nil }
func (db *DatabaseMock) GetFileLabels(colonyName string) ([]*core.Label, error) { return nil, nil }
func (db *DatabaseMock) GetFileLabelsByName(colonyName string, name string, exact bool) ([]*core.Label, error) { return nil, nil }
func (db *DatabaseMock) GetFilesByColonyName(colonyName string) ([]*core.File, error) { return nil, nil }
func (db *DatabaseMock) GetFilesByProcessGraphID(processGraphID string) ([]*core.File, error) { return nil, nil }
func (db *DatabaseMock) GetFiles() ([]*core.File, error) { return nil, nil }
func (db *DatabaseMock) UpdateFile(file *core.File) error { return nil }
func (db *DatabaseMock) RemoveFileByID(colonyName string, fileID string) error { return nil }
func (db *DatabaseMock) RemoveFileByName(colonyName string, label string, name string) error { return nil }
func (db *DatabaseMock) CountFiles(colonyName string) (int, error) { return 0, nil }
func (db *DatabaseMock) CountFilesWithLabel(colonyName string, label string) (int, error) { return 0, nil }


// SnapshotDatabase interface
func (db *DatabaseMock) CreateSnapshot(colonyName string, label string, name string) (*core.Snapshot, error) { return nil, nil }
func (db *DatabaseMock) GetSnapshotByID(colonyName string, snapshotID string) (*core.Snapshot, error) { return nil, nil }
func (db *DatabaseMock) GetSnapshotsByColonyName(colonyName string) ([]*core.Snapshot, error) { return nil, nil }
func (db *DatabaseMock) RemoveSnapshotByID(colonyName string, snapshotID string) error { return nil }
func (db *DatabaseMock) GetSnapshotByName(colonyName string, name string) (*core.Snapshot, error) { return nil, nil }
func (db *DatabaseMock) RemoveSnapshotByName(colonyName string, name string) error { return nil }
func (db *DatabaseMock) RemoveSnapshotsByColonyName(colonyName string) error { return nil }
func (db *DatabaseMock) CountSnapshots() (int, error) { return 0, nil }


// SecurityDatabase interface
func (db *DatabaseMock) SetServerID(oldServerID, newServerID string) error { return nil }
func (db *DatabaseMock) GetServerID() (string, error) { return "", nil }
func (db *DatabaseMock) ChangeColonyID(colonyName string, oldColonyID, newColonyID string) error { return nil }
func (db *DatabaseMock) ChangeUserID(colonyName string, oldUserID, newUserID string) error { return nil }
func (db *DatabaseMock) ChangeExecutorID(colonyName string, oldExecutorID, newExecutorID string) error { return nil }

// Implement the database.Database interface
func (db *DatabaseMock) CreateTables() error { return nil }
func (db *DatabaseMock) DropTables() error { return nil }
func (db *DatabaseMock) Close()             { }
func (db *DatabaseMock) Initialize() error { return nil }
func (db *DatabaseMock) Drop() error { return nil }
func (db *DatabaseMock) Lock(timeout int) error { return nil }
func (db *DatabaseMock) Unlock() error { return nil }
func (db *DatabaseMock) ApplyRetentionPolicy(retentionPeriod int64) error { return nil }

// Test utility functions
func createFakeColoniesController() (*ColoniesController, *DatabaseMock) {
	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: constants.TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	dbMock := &DatabaseMock{}
	return CreateColoniesController(dbMock, node, clusterConfig, "/tmp/colonies/etcd", constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, false, -1, 500), dbMock
}

func createTestColoniesController(db *postgresql.PQDatabase) *ColoniesController {
	node := cluster.Node{Name: "test", Host: "localhost", EtcdClientPort: 24101, EtcdPeerPort: 23101, RelayPort: 25101, APIPort: constants.TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	return CreateColoniesController(db, node, clusterConfig, "/tmp/colonies/etcd_test", constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, false, -1, 500)
}

func createTestColoniesController2(db *postgresql.PQDatabase) *ColoniesController {
	node := cluster.Node{Name: "test2", Host: "localhost", EtcdClientPort: 24102, EtcdPeerPort: 23102, RelayPort: 25102, APIPort: constants.TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	return CreateColoniesController(db, node, clusterConfig, "/tmp/colonies/etcd_test2", constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, false, -1, 500)
}